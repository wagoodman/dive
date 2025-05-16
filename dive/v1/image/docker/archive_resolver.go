package docker

import (
	"context"
	"fmt"
	"github.com/wagoodman/dive/dive/v1/image"
	"os"
)

type archiveResolver struct{}

func NewResolverFromArchive() *archiveResolver {
	return &archiveResolver{}
}

// Name returns the name of the resolver to display to the user.
func (r *archiveResolver) Name() string {
	return "docker-archive"
}

func (r *archiveResolver) Fetch(ctx context.Context, path string) (*image.Image, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	img, err := NewImageArchive(reader)
	if err != nil {
		return nil, err
	}
	return img.ToImage(path)
}

func (r *archiveResolver) Build(ctx context.Context, args []string) (*image.Image, error) {
	return nil, fmt.Errorf("build option not supported for docker archive resolver")
}

func (r *archiveResolver) Extract(ctx context.Context, id string, l string, p string) error {
	return fmt.Errorf("not implemented")
}
