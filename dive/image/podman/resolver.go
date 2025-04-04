//go:build linux || darwin
// +build linux darwin

package podman

import (
	"context"
	"fmt"
	"io"

	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/dive/image/docker"
)

type resolver struct{}

func NewResolverFromEngine() *resolver {
	return &resolver{}
}

// Name returns the name of the resolver to display to the user.
func (r *resolver) Name() string {
	return "podman"
}

func (r *resolver) Build(ctx context.Context, args []string) (*image.Image, error) {
	id, err := buildImageFromCli(args)
	if err != nil {
		return nil, err
	}
	return r.Fetch(ctx, id)
}

func (r *resolver) Fetch(ctx context.Context, id string) (*image.Image, error) {
	// todo: add podman fetch attempt via varlink first...

	img, err := r.resolveFromDockerArchive(id)
	if err == nil {
		return img, err
	}

	return nil, fmt.Errorf("unable to resolve image '%s': %+v", id, err)
}

func (r *resolver) Extract(ctx context.Context, id string, l string, p string) error {
	// todo: add podman fetch attempt via varlink first...

	err, reader := streamPodmanCmd("image", "save", id)
	if err != nil {
		return err
	}

	if err := docker.ExtractFromImage(io.NopCloser(reader), l, p); err == nil {
		return nil
	}

	return fmt.Errorf("unable to extract from image '%s': %+v", id, err)
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
	return img.ToImage(id)
}
