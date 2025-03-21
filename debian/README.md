
## Guides
- https://go-team.pages.debian.net/packaging.html
- https://www.debian.org/doc/manuals/debmake-doc/ch04.en.html

## Tools
- generate debian dir `dh-make-golang make  -type p github.com/tonymet/dive`
- create chroot (lots of deps)`debootstrap bullseye /srv/chroot/test`
- build package - `dpkg-buildpackage -b`
- [in schroot] install deps `mk-build-deps -i`

## TODO
- address missing deps (some are not in deb)
- copyright & upstream updates to canonical


## Missing deps

```
 /usr/lib/go-1.15/src/github.com/wagoodman/dive/cmd (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/cmd (from $GOPATH)
src/github.com/tonymet/dive/cmd/analyze.go:11:2: cannot find package "github.com/wagoodman/dive/dive" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/dive (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/dive (from $GOPATH)
src/github.com/tonymet/dive/cmd/root.go:16:2: cannot find package "github.com/wagoodman/dive/dive/filetree" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/dive/filetree (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/dive/filetree (from $GOPATH)
src/github.com/tonymet/dive/cmd/analyze.go:12:2: cannot find package "github.com/wagoodman/dive/runtime" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/runtime (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/runtime (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/dive/get_image_resolver.go:5:2: cannot find package "github.com/wagoodman/dive/dive/image" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/dive/image (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/dive/image (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/dive/get_image_resolver.go:6:2: cannot find package "github.com/wagoodman/dive/dive/image/docker" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/dive/image/docker (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/dive/image/docker (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/dive/get_image_resolver.go:7:2: cannot find package "github.com/wagoodman/dive/dive/image/podman" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/dive/image/podman (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/dive/image/podman (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/dive/filetree/file_node.go:13:2: cannot find package "github.com/phayes/permbits" in any of:
        /usr/lib/go-1.15/src/github.com/phayes/permbits (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/phayes/permbits (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/dive/image/docker/engine_resolver.go:11:2: cannot find package "github.com/docker/cli/cli/connhelper" in any of:
        /usr/lib/go-1.15/src/github.com/docker/cli/cli/connhelper (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/docker/cli/cli/connhelper (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/dive/image/docker/cli.go:5:2: cannot find package "github.com/wagoodman/dive/utils" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/utils (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/utils (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/runtime/run.go:11:2: cannot find package "github.com/wagoodman/dive/runtime/ci" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/runtime/ci (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/runtime/ci (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/runtime/run.go:12:2: cannot find package "github.com/wagoodman/dive/runtime/export" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/runtime/export (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/runtime/export (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/runtime/run.go:13:2: cannot find package "github.com/wagoodman/dive/runtime/ui" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/runtime/ui (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/runtime/ui (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/runtime/ui/app.go:10:2: cannot find package "github.com/jroimartin/gocui" in any of:
        /usr/lib/go-1.15/src/github.com/jroimartin/gocui (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/jroimartin/gocui (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/runtime/ui/app.go:5:2: cannot find package "github.com/wagoodman/dive/runtime/ui/key" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/runtime/ui/key (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/runtime/ui/key (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/runtime/ui/app.go:6:2: cannot find package "github.com/wagoodman/dive/runtime/ui/layout" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/runtime/ui/layout (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/runtime/ui/layout (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/runtime/ui/app.go:7:2: cannot find package "github.com/wagoodman/dive/runtime/ui/layout/compound" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/runtime/ui/layout/compound (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/runtime/ui/layout/compound (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/runtime/ui/controller.go:8:2: cannot find package "github.com/wagoodman/dive/runtime/ui/view" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/runtime/ui/view (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/runtime/ui/view (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/runtime/ui/controller.go:9:2: cannot find package "github.com/wagoodman/dive/runtime/ui/viewmodel" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/runtime/ui/viewmodel (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/runtime/ui/viewmodel (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/runtime/ui/key/binding.go:8:2: cannot find package "github.com/wagoodman/dive/runtime/ui/format" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/dive/runtime/ui/format (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/dive/runtime/ui/format (from $GOPATH)
src/github.com/tonymet/dive/deb/dive/runtime/ui/key/binding.go:9:2: cannot find package "github.com/wagoodman/keybinding" in any of:
        /usr/lib/go-1.15/src/github.com/wagoodman/keybinding (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/wagoodman/keybinding (from $GOPATH)
src/github.com/tonymet/dive/runtime/ui/app.go:6:2: cannot find package "github.com/awesome-gocui/gocui" in any of:
        /usr/lib/go-1.15/src/github.com/awesome-gocui/gocui (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/awesome-gocui/gocui (from $GOPATH)
src/github.com/tonymet/dive/runtime/ui/key/binding.go:7:2: cannot find package "github.com/awesome-gocui/keybinding" in any of:
        /usr/lib/go-1.15/src/github.com/awesome-gocui/keybinding (from $GOROOT)
        /home/tonymet/src/dive/_build/src/github.com/awesome-gocui/keybinding (from $GOPATH)
```