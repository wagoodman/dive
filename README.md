# dive
[![GitHub release](https://img.shields.io/github/release/wagoodman/dive.svg)](https://github.com/wagoodman/dive/releases/latest)
[![Validations](https://github.com/wagoodman/dive/actions/workflows/validations.yaml/badge.svg)](https://github.com/wagoodman/dive/actions/workflows/validations.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/wagoodman/dive)](https://goreportcard.com/report/github.com/wagoodman/dive)
[![License: MIT](https://img.shields.io/badge/License-MIT%202.0-blue.svg)](https://github.com/wagoodman/dive/blob/main/LICENSE)
[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg?style=flat)](https://www.paypal.me/wagoodman)

**A tool for exploring a Docker image, layer contents, and discovering ways to shrink the size of your Docker/OCI image.**


![Image](.data/demo.gif)

To analyze a Docker image simply run dive with an image tag/id/digest:
```bash
dive <your-image-tag>
```

or you can dive with Docker directly:
```
alias dive="docker run -ti --rm  -v /var/run/docker.sock:/var/run/docker.sock docker.io/wagoodman/dive"
dive <your-image-tag>

# for example
dive nginx:latest
```

or if you want to build your image then jump straight into analyzing it:
```bash
dive build -t <some-tag> .
```

Building on macOS (supporting only the Docker container engine):

```bash
docker run --rm -it \
      -v /var/run/docker.sock:/var/run/docker.sock \
      -v  "$(pwd)":"$(pwd)" \
      -w "$(pwd)" \
      -v "$HOME/.dive.yaml":"$HOME/.dive.yaml" \
      docker.io/wagoodman/dive:latest build -t <some-tag> .
```

Additionally you can run this in your CI pipeline to ensure you're keeping wasted space to a minimum (this skips the UI):
```
CI=true dive <your-image>
```

![Image](.data/demo-ci.png)

**This is beta quality!** *Feel free to submit an issue if you want a new feature or find a bug :)*

## Basic Features

**Show Docker image contents broken down by layer**

As you select a layer on the left, you are shown the contents of that layer combined with all previous layers on the right. Also, you can fully explore the file tree with the arrow keys.

**Indicate what's changed in each layer**

Files that have changed, been modified, added, or removed are indicated in the file tree. This can be adjusted to show changes for a specific layer, or aggregated changes up to this layer.

**Estimate "image efficiency"**

The lower left pane shows basic layer info and an experimental metric that will guess how much wasted space your image contains. This might be from duplicating files across layers, moving files across layers, or not fully removing files. Both a percentage "score" and total wasted file space is provided.

**Quick build/analysis cycles**

You can build a Docker image and do an immediate analysis with one command:
`dive build -t some-tag .`

You only need to replace your `docker build` command with the same `dive build`
command.

**CI Integration**

Analyze an image and get a pass/fail result based on the image efficiency and wasted space. Simply set `CI=true` in the environment when invoking any valid dive command.

**Multiple Image Sources and Container Engines Supported**

With the `--source` option, you can select where to fetch the container image from:
```bash
dive <your-image> --source <source>
```
or
```bash
dive <source>://<your-image>
```

With valid `source` options as such:
- `docker`: Docker engine (the default option)
- `docker-archive`: A Docker Tar Archive from disk
- `podman`: Podman engine (linux only)

## Installation

**Ubuntu/Debian**

Using debs:
```bash
DIVE_VERSION=$(curl -sL "https://api.github.com/repos/wagoodman/dive/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
curl -fOL "https://github.com/wagoodman/dive/releases/download/v${DIVE_VERSION}/dive_${DIVE_VERSION}_linux_amd64.deb"
sudo apt install ./dive_${DIVE_VERSION}_linux_amd64.deb
```

Using snap:
```bash
sudo snap install docker
sudo snap install dive
sudo snap connect dive:docker-executables docker:docker-executables
sudo snap connect dive:docker-daemon docker:docker-daemon
```

> [!CAUTION]
> The Snap method is not recommended if you installed Docker via `apt-get`, since it might break your existing Docker daemon.
> 
> See also: https://github.com/wagoodman/dive/issues/546


**RHEL/Centos**
```bash
DIVE_VERSION=$(curl -sL "https://api.github.com/repos/wagoodman/dive/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
curl -fOL "https://github.com/wagoodman/dive/releases/download/v${DIVE_VERSION}/dive_${DIVE_VERSION}_linux_amd64.rpm"
rpm -i dive_${DIVE_VERSION}_linux_amd64.rpm
```

**Arch Linux**

Available in the [extra repository](https://archlinux.org/packages/extra/x86_64/dive/) and can be installed via [pacman](https://wiki.archlinux.org/title/Pacman):

```bash
pacman -S dive
```

**Mac**

If you use [Homebrew](https://brew.sh):

```bash
brew install dive
```

If you use [MacPorts](https://www.macports.org):

```bash
sudo port install dive
```

Or download the latest Darwin build from the [releases page](https://github.com/wagoodman/dive/releases/latest).

**Windows**

If you use [Chocolatey](https://chocolatey.org)

```powershell
choco install dive
```

If you use [scoop](https://scoop.sh/)

```powershell
scoop install main/dive
```

If you use [winget](https://learn.microsoft.com/en-gb/windows/package-manager/):

```powershell
winget install --id wagoodman.dive
```

Or download the latest Windows build from the [releases page](https://github.com/wagoodman/dive/releases/latest).

**Go tools**
Requires Go version 1.10 or higher.

```bash
go install github.com/wagoodman/dive@latest
```
*Note*: installing in this way you will not see a proper version when running `dive -v`.

**Nix/NixOS**

On NixOS:
```bash
nix-env -iA nixos.dive
```
On non-NixOS (Linux, Mac)
```bash
nix-env -iA nixpkgs.dive
```

**X-CMD**

[x-cmd](https://www.x-cmd.com/) is a **toolbox for Posix Shell**, offering a lightweight package manager built using shell and awk.
```sh
x env use dive
```

**Docker**
```bash
docker pull docker.io/wagoodman/dive
# or alternatively
docker pull ghcr.io/wagoodman/dive
```

When running you'll need to include the Docker socket file:
```bash
docker run --rm -it \
    -v /var/run/docker.sock:/var/run/docker.sock \
    docker.io/wagoodman/dive:latest <dive arguments...>
```

Docker for Windows (showing PowerShell compatible line breaks; collapse to a single line for Command Prompt compatibility)
```bash
docker run --rm -it `
    -v /var/run/docker.sock:/var/run/docker.sock `
    docker.io/wagoodman/dive:latest <dive arguments...>
```

**Note:** depending on the version of docker you are running locally you may need to specify the docker API version as an environment variable:
```bash
   DOCKER_API_VERSION=1.37 dive ...
```
or if you are running with a docker image:
```bash
docker run --rm -it \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -e DOCKER_API_VERSION=1.37 \
    docker.io/wagoodman/dive:latest <dive arguments...>
```
if you are using an alternative runtime (Colima etc) then you may need to specify the docker host as an environment variable in order to pull local images:
```bash
   export DOCKER_HOST=$(docker context inspect -f '{{ .Endpoints.docker.Host }}')
```

## CI Integration

When running dive with the environment variable `CI=true` then the dive UI will be bypassed and will instead analyze your docker image, giving it a pass/fail indication via return code. Currently there are three metrics supported via a `.dive-ci` file that you can put at the root of your repo:
```
rules:
  # If the efficiency is measured below X%, mark as failed.
  # Expressed as a ratio between 0-1.
  lowestEfficiency: 0.95

  # If the amount of wasted space is at least X or larger than X, mark as failed.
  # Expressed in B, KB, MB, and GB.
  highestWastedBytes: 20MB

  # If the amount of wasted space makes up for X% or more of the image, mark as failed.
  # Note: the base image layer is NOT included in the total image size.
  # Expressed as a ratio between 0-1; fails if the threshold is met or crossed.
  highestUserWastedPercent: 0.20
```
You can override the CI config path with the `--ci-config` option.

## KeyBindings

Key Binding                                | Description
-------------------------------------------|---------------------------------------------------------
<kbd>Ctrl + C</kbd> or <kbd>Q</kbd>        | Exit
<kbd>Tab</kbd>                             | Switch between the layer and filetree views
<kbd>Ctrl + F</kbd>                        | Filter files
<kbd>ESC</kbd>                             | Close filter files
<kbd>PageUp</kbd> or <kbd>U</kbd>          | Scroll up a page
<kbd>PageDown</kbd> or <kbd>D</kbd>        | Scroll down a page
<kbd>Up</kbd> or <kbd>K</kbd>              | Move up one line within a page
<kbd>Down</kbd> or <kbd>J</kbd>            | Move down one line within a page
<kbd>Ctrl + A</kbd>                        | Layer view: see aggregated image modifications
<kbd>Ctrl + L</kbd>                        | Layer view: see current layer modifications
<kbd>Space</kbd>                           | Filetree view: collapse/uncollapse a directory
<kbd>Ctrl + Space</kbd>                    | Filetree view: collapse/uncollapse all directories
<kbd>Ctrl + A</kbd>                        | Filetree view: show/hide added files
<kbd>Ctrl + R</kbd>                        | Filetree view: show/hide removed files
<kbd>Ctrl + M</kbd>                        | Filetree view: show/hide modified files
<kbd>Ctrl + U</kbd>                        | Filetree view: show/hide unmodified files
<kbd>Ctrl + B</kbd>                        | Filetree view: show/hide file attributes
<kbd>PageUp</kbd> or <kbd>U</kbd>          | Filetree view: scroll up a page
<kbd>PageDown</kbd> or <kbd>D</kbd>        | Filetree view: scroll down a page

## UI Configuration

No configuration is necessary, however, you can create a config file and override values:
```yaml
# supported options are "docker" and "podman"
container-engine: docker
# continue with analysis even if there are errors parsing the image archive
ignore-errors: false
log:
  enabled: true
  path: ./dive.log
  level: info

# Note: you can specify multiple bindings by separating values with a comma.
# Note: UI hinting is derived from the first binding
keybinding:
  # Global bindings
  quit: ctrl+c
  toggle-view: tab
  filter-files: ctrl+f, ctrl+slash
  close-filter-files: esc
  up: up,k
  down: down,j
  left: left,h
  right: right,l

  # Layer view specific bindings
  compare-all: ctrl+a
  compare-layer: ctrl+l

  # File view specific bindings
  toggle-collapse-dir: space
  toggle-collapse-all-dir: ctrl+space
  toggle-added-files: ctrl+a
  toggle-removed-files: ctrl+r
  toggle-modified-files: ctrl+m
  toggle-unmodified-files: ctrl+u
  toggle-filetree-attributes: ctrl+b
  page-up: pgup,u
  page-down: pgdn,d

diff:
  # You can change the default files shown in the filetree (right pane). All diff types are shown by default.
  hide:
    - added
    - removed
    - modified
    - unmodified

filetree:
  # The default directory-collapse state
  collapse-dir: false

  # The percentage of screen width the filetree should take on the screen (must be >0 and <1)
  pane-width: 0.5

  # Show the file attributes next to the filetree
  show-attributes: true

layer:
  # Enable showing all changes from this layer and every previous layer
  show-aggregated-changes: false

```

dive will search for configs in the following locations:
- `$XDG_CONFIG_HOME/dive/*.yaml`
- `$XDG_CONFIG_DIRS/dive/*.yaml`
- `~/.config/dive/*.yaml`
- `~/.dive.yaml`

`.yml` can be used instead of `.yaml` if desired.
