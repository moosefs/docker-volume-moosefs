# Docker Volume Plugin for MooseFS

Plugin for MooseFS to create persistent volumes in Docker containers.

<img align="right" alt="MooseFS logo" src="https://moosefs.com/Content/Images/moosefs.png" />

## Version history:

- 0.2.0 Creates new subdirectory for each volume in root e.g. /mnt/moosefs
- 0.1.0 Mountpoint /mnt/moosefs is used as the volume

## Preconditions

- **MooseFS** Storage Cluster has to be setup and running
- **MooseFS** Client should be installed on host machine
- **MooseFS** is mounted in one of host directories

## Installation

We provide pre-built binary **rpm** and **deb** packages available from the [releases](https://github.com/moosefs/docker-volume-moosefs/releases) page. You can download them and go to **installation** for your OS.

### RedHat/CentOS

Pre-built **rpm** package can be downloaded from [releases](https://github.com/moosefs/docker-volume-moosefs/releases) or can be built with following commands:

```
make rpm-deps
make
make rpm
```

**Installation**

Install and start the service:

```
yum localinstall docker-volume-moosefs-$VERSION.rpm
systemctl start docker-volume-moosefs
```

**Deb** package can be also build using following command:

```
make deb
```

### Debian

Pre-built **deb** package can be downloaded from [releases](https://github.com/moosefs/docker-volume-moosefs/releases) or can be built with following commands:

```
make deb-deps
make
make deb
```

**Installation**

Install and start the service:

```
dpkg -i docker-volume-moosefs_$VERSION.deb
systemctl start docker-volume-moosefs
```

**Rpm** package can be also build using following command:

```
make rpm
```

## Usage example

Assuming we have **MooseFS** mounted in `/mnt/moosefs` we will create a volume labeled `mymoosefs`:

```
docker volume create -d moosefs --name mymoosefs -o root=/mnt/moosefs
```

Without specified mountpoint plugin will assume mounting in `/mnt/moosefs/$NAME`, so **mymoosefs** will be mounted in `/mnt/moosefs/mymoosefs`

We can use shorter command: `docker volume create -d moosefs --name mymoosefs`

We can inspect created volume with following command:

```
docker volume inspect mymoosefs
```

Now we can use our `mymoosefs` **MooseFS Volume** in example container such as Nginx. Following commang will mount our storage to `/usr/share/nginx/html` directory, where nginx stores html files. We are forwarding port 80 from container to [http://localhost:10080](http://localhost:10080). This command will start Nginx in container:

```
docker run -ti -v mymoosefs:/usr/share/nginx/html -p 10080:80 nginx:latest bash -c "service nginx start; bash"
```

Now we can check that Nginx created two html files in `/mnt/moosefs`:

```
ls /mnt/moosefs
  50x.html  index.html
```

These file will still exist in MooseFS, after shutting down the container or removing volume.


To remove the volume (note that this will ***NOT*** remove the actual data):

```
docker volume rm mymoosefs
```
