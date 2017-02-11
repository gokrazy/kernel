# gokrazy kernel repository

This repository holds a pre-built Linux kernel image for the Raspberry
Pi 3, used by the [gokrazy](https://github.com/gokrazy/gokrazy)
project.

The files in this repository are picked up automatically by
`gokr-packer`, so you don’t need to interact with this repository
unless you want to update the kernel to a custom version.

## Updating the kernel

First, follow the [gokrazy installation instructions](https://github.com/gokrazy/gokrazy).

We’re using docker to get a reproducible build environment for our
kernel images, so install docker if you haven’t already:
```
sudo apt install docker.io
sudo addgroup $USER docker
newgrp docker
```

Install the kernel-related gokrazy tools:
```
go install github.com/gokrazy/kernel/cmd/...
```

And build a new kernel (takes about 5 minutes):
```
gokr-rebuild-kernel
```

The new kernel is stored in `$GOPATH/src/github.com/gokrazy/kernel` so
that it will be picked up by `gokr-packer`.