package image

import "golang.org/x/net/context"

type Resolver interface {
	Name() string
	Fetch(ctx context.Context, id string) (*Image, error)
	Build(ctx context.Context, options []string) (*Image, error)
	ContentReader
}

type ContentReader interface {
	Extract(ctx context.Context, id string, layer string, path string) error
}
