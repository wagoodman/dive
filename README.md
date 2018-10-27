# dive
[![Go Report Card](https://goreportcard.com/badge/github.com/wagoodman/dive)](https://goreportcard.com/report/github.com/wagoodman/dive)

**A tool for exploring a docker image, layer contents, and discovering ways to shrink your Docker image size.**

![Image](.data/demo.gif)

To analyze a Docker image simply run dive with an image tag/id/digest:
```bash
dive <your-image-tag>
```

or if you want to build your image then jump straight into analyzing it:
```bash
dive build -t <some-tag> .
```

**This is beta quality!** *Feel free to submit an issue if you want a new feature or find a bug :)*

## Basic Features

**Show Docker image contents broken down by layer**

As you select a layer on the left, you are shown the contents of that layer
combined with all previous layers on the right. Also, you can fully explore the
file tree with the arrow keys.

**Indicate what's changed in each layer**

Files that have changed, been modified, added, or removed are indicated in the
file tree. This can be adjusted to show changes for a specific layer, or
aggregated changes up to this layer.

**Estimate "image efficiency"**

The lower left pane shows basic layer info and an experimental metric that will
guess how much wasted space your image contains. This might be from duplicating
files across layers, moving files across layers, or not fully removing files.
Both a percentage "score" and total wasted file space is provided.

**Quick build/analysis cycles**

You can build a Docker image and do an immediate analysis with one command:
`dive build -t some-tag .`

You only need to replace your `docker build` command with the same `dive build`
command.


## Installation

**Ubuntu/Debian**
```bash
wget https://github.com/wagoodman/dive/releases/download/v0.0.8/dive_0.0.8_linux_amd64.deb
sudo apt install ./dive_0.0.8_linux_amd64.deb
```

**RHEL/Centos**
```bash
wget https://github.com/wagoodman/dive/releases/download/v0.0.8/dive_0.0.8_linux_amd64.rpm
rpm -i dive_0.0.8_linux_amd64.rpm
```

**Mac**
```bash
brew tap wagoodman/dive
brew install dive
```
or download a Darwin build from the releases page.

**Go tools**
```bash
go get github.com/wagoodman/dive
```

**Docker**
```bash
docker pull wagoodman/dive
```

or 

```bash
docker pull quay.io/wagoodman/dive
```

When running you'll need to include the docker client binary and socket file:
```bash
docker run --rm -it \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v $(which docker):/bin/docker \
    wagoodman/dive:latest <dive arguments...>
```
For docker in windows (does not support pulling images yet):
```bash
docker run --rm -it -v //var/run/docker.sock:/var/run/docker.sock wagoodman/dive:latest <dive arguments...>
```

