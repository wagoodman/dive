//go:build !linux && !darwin
// +build !linux,!darwin

package podman

import (
	"context"
	"fmt"
	"github.com/wagoodman/dive/dive/v1/image"
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
	return nil, fmt.Errorf("unsupported platform")
}

func (r *resolver) Fetch(ctx context.Context, id string) (*image.Image, error) {
	return nil, fmt.Errorf("unsupported platform")
}

func (r *resolver) Extract(ctx context.Context, id string, l string, p string) error {
	return fmt.Errorf("unsupported platform")
}
