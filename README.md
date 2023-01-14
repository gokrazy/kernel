# gokrazy kernel repository

This repository holds a pre-built Linux kernel image for the Raspberry Pi 3, Pi
4, and Pi Zero 2 W, used by the [gokrazy](https://gokrazy.org/) project.

The files in this repository are picked up automatically by
the `gok` tool, so you don’t need to interact with this repository
unless you want to update the kernel to a custom version.

## Cloning the kernel repository

This repository clocks in at over 3 GB of disk usage, so you might want to clone
it as a shallow clone:

```
git clone --depth=1 https://github.com/gokrazy/kernel
```

## Updating the kernel

First, follow the [gokrazy installation instructions](https://gokrazy.org/quickstart/).

We’re using docker to get a reproducible build environment for our
kernel images, so install docker if you haven’t already:
```
sudo apt install docker.io
sudo addgroup $USER docker
newgrp docker
```

Clone the kernel git repository:
```
git clone --depth=1 https://github.com/gokrazy/kernel
cd kernel
```

Install the kernel-related gokrazy tools:
```
go install ./cmd/...
```


And build a new kernel (takes about 5 minutes):
```
gokr-rebuild-kernel
```

The new kernel is stored in the working directory. Use `gok add .` to
ensure the next `gok` build will pick up your changed files.
