package main

import (
	"errors"
	"fmt"
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
	name       string
	mountpoint string
}

type moosefsDriver struct {
	mounts map[string]*moosefsMount
	m      *sync.Mutex
}

func newMooseFSDriver(mountpoint string) moosefsDriver {
	d := moosefsDriver{
		mounts: make(map[string]*moosefsMount),
		m:      &sync.Mutex{},
	}
	return d
}

func (d moosefsDriver) Create(r *volume.CreateRequest) error {
	var volumeMountpoint string

	d.m.Lock()
	defer d.m.Unlock()

	if optsMountpoint, ok := r.Options["mountpoint"]; ok {
		volumeMountpoint = optsMountpoint
	} else {
		// Assume the default mountpoint
		volumeMountpoint = filepath.Join(*mountpoint, r.Name)
	}

	if !ismoosefs(volumeMountpoint) {
		emsg := fmt.Sprintf("Cannot create volume %s as it's not a valid MooseFS mount", volumeMountpoint)
		log.Error(emsg)
		return errors.New(emsg)
	}

	if _, ok := d.mounts[r.Name]; ok {
		emsg := fmt.Sprintf("Cannot create volume %s, it already exists", volumeMountpoint)
		log.Error(emsg)
		return errors.New(emsg)
	}

	d.mounts[r.Name] = &moosefsMount{
		name:       r.Name,
		mountpoint: volumeMountpoint,
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
		return &volume.PathResponse{Mountpoint: d.mounts[r.Name].mountpoint}, nil
	}
	return &volume.PathResponse{}, nil
}

func (d moosefsDriver) Mount(r *volume.MountRequest) (*volume.MountResponse, error) {
	volumeMountpoint := d.mounts[r.Name].mountpoint
	if !ismoosefs(volumeMountpoint) {
		emsg := fmt.Sprintf("Cannot mount volume %s as it's not a valid MooseFS mount", volumeMountpoint)
		log.Error(emsg)
		return &volume.MountResponse{}, errors.New(emsg)
	}
	if _, ok := d.mounts[r.Name]; ok {
		return &volume.MountResponse{Mountpoint: d.mounts[r.Name].mountpoint}, nil
	}
	return &volume.MountResponse{}, nil
}

func (d moosefsDriver) Unmount(r *volume.UnmountRequest) error {
	return nil
}

func (d moosefsDriver) Get(r *volume.GetRequest) (*volume.GetResponse, error) {
	if v, ok := d.mounts[r.Name]; ok {
		return &volume.GetResponse{
			Volume: &volume.Volume{Name: v.name, Mountpoint: v.mountpoint}}, nil
	}
	return &volume.GetResponse{}, fmt.Errorf("volume %s unknown", r.Name)
}

func (d moosefsDriver) List() (*volume.ListResponse, error) {
	volumes := []*volume.Volume{}
	for v := range d.mounts {
		volumes = append(volumes, &volume.Volume{Name: d.mounts[v].name, Mountpoint: d.mounts[v].mountpoint})
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
