//go:build linux || darwin
// +build linux darwin

package podman

import (
	"fmt"
	"io"

	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/dive/image/docker"
)

type resolver struct{}

func NewResolverFromEngine() *resolver {
	return &resolver{}
}

func (r *resolver) Build(args []string) (*image.Image, error) {
	id, err := buildImageFromCli(args)
	if err != nil {
		return nil, err
	}
	return r.Fetch(id)
}

func (r *resolver) Fetch(id string) (*image.Image, error) {
	// todo: add podman fetch attempt via varlink first...

	img, err := r.resolveFromDockerArchive(id)
	if err == nil {
		return img, err
	}

	return nil, fmt.Errorf("unable to resolve image '%s': %+v", id, err)
}

func (r *resolver) resolveFromDockerArchive(id string) (*image.Image, error) {
	err, reader := streamPodmanCmd("image", "save", id)
	if err != nil {
		return nil, err
	}

	img, err := docker.NewImageArchive(io.NopCloser(reader))
	if err != nil {
		return nil, err
	}
	return img.ToImage()
}
