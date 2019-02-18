package main

import (
	"errors"
	"fmt"
    "os"
	"path"
	"path/filepath"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/davecgh/go-spew/spew"
	"github.com/docker/go-plugins-helpers/volume"
)

// A single volume instance
type moosefsMount struct {
	name    string
    path    string
	root    string
}

type moosefsDriver struct {
	mounts map[string]*moosefsMount
	m      *sync.Mutex
}

func newMooseFSDriver(root string) moosefsDriver {
	d := moosefsDriver{
		mounts: make(map[string]*moosefsMount),
		m:      &sync.Mutex{},
	}
	return d
}

func (d moosefsDriver) Create(r *volume.CreateRequest) error {
    var volumeRoot string

	d.m.Lock()
	defer d.m.Unlock()

	if optsRoot, ok := r.Options["root"]; ok {
		volumeRoot = optsRoot
	} else {
		// Assume the default root
        volumeRoot = *root
	}

    volumePath := filepath.Join(volumeRoot, r.Name)

    if err := mkdir(volumePath); err != nil {
		return err
	}
    
	if !ismoosefs(volumePath) {
		emsg := fmt.Sprintf("Cannot create volume %s as it's not a valid MooseFS mount", volumePath)
		log.Error(emsg)
		return errors.New(emsg)
	}

	if _, ok := d.mounts[r.Name]; ok {
		emsg := fmt.Sprintf("Cannot create volume %s, it already exists", volumePath)
		log.Error(emsg)
		return errors.New(emsg)
	}

    if err := mkdir(volumePath); err != nil {
		return err
	}
	d.mounts[r.Name] = &moosefsMount{
		name:   r.Name,
        path:   volumePath,
		root:   volumeRoot,
	}

	if *verbose {
		spew.Dump(d.mounts)
	}

	return nil
}

func (d moosefsDriver) Remove(r *volume.RemoveRequest) error {
	d.m.Lock()
	defer d.m.Unlock()
	if _, ok := d.mounts[r.Name]; ok {
		delete(d.mounts, r.Name)
	}
	return nil
}

func (d moosefsDriver) Path(r *volume.PathRequest) (*volume.PathResponse, error) {
	if _, ok := d.mounts[r.Name]; ok {
		return &volume.PathResponse{Mountpoint: d.mounts[r.Name].path}, nil
	}
	return &volume.PathResponse{}, nil
}

func (d moosefsDriver) Mount(r *volume.MountRequest) (*volume.MountResponse, error) {
	volumePath := filepath.Join(d.mounts[r.Name].root, r.Name)
	if !ismoosefs(volumePath) {
		emsg := fmt.Sprintf("Cannot mount volume %s as it's not a valid MooseFS mount", volumePath)
		log.Error(emsg)
		return &volume.MountResponse{}, errors.New(emsg)
	}
	if _, ok := d.mounts[r.Name]; ok {
		return &volume.MountResponse{Mountpoint: d.mounts[r.Name].path}, nil
	}
	return &volume.MountResponse{}, nil
}

func (d moosefsDriver) Unmount(r *volume.UnmountRequest) error {
	return nil
}

func (d moosefsDriver) Get(r *volume.GetRequest) (*volume.GetResponse, error) {
	if v, ok := d.mounts[r.Name]; ok {
		return &volume.GetResponse{
			Volume: &volume.Volume{Name: v.name, Mountpoint: v.path}}, nil
	}
	return &volume.GetResponse{}, fmt.Errorf("volume %s unknown", r.Name)
}

func (d moosefsDriver) List() (*volume.ListResponse, error) {
	volumes := []*volume.Volume{}
	for v := range d.mounts {
		volumes = append(volumes, &volume.Volume{Name: d.mounts[v].name, Mountpoint: d.mounts[v].path})
	}
	return &volume.ListResponse{Volumes: volumes}, nil
}

func (d moosefsDriver) Capabilities() *volume.CapabilitiesResponse {
	var res volume.CapabilitiesResponse
	res.Capabilities = volume.Capability{Scope: "global"}
	return &res
}

// Check if MooseFS is mounted in mountpoint using the .masterinfo file
func ismoosefs(mountpoint string) bool {
	stat := syscall.Statfs_t{}
	err := syscall.Statfs(path.Join(mountpoint, ".masterinfo"), &stat)
	if err != nil {
		log.Errorf("Could not determine filesystem type for %s: %s", mountpoint, err)
		return false
	}
	return true
}

func mkdir(path string) error {
	fstat, err := os.Lstat(path)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if fstat != nil && !fstat.IsDir() {
		return fmt.Errorf("%v already exist and it's not a directory", path)
	}

	return nil
}
