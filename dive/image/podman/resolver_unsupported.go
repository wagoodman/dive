//go:build !linux && !darwin
// +build !linux,!darwin

package podman

import (
	"fmt"

	"github.com/joschi/dive/dive/image"
)

type resolver struct{}

func NewResolverFromEngine() *resolver {
	return &resolver{}
}

// Name returns the name of the resolver to display to the user.
func (r *resolver) Name() string {
	return "podman"
}
func (r *resolver) Build(args []string) (*image.Image, error) {
	return nil, fmt.Errorf("unsupported platform")
}

func (r *resolver) Fetch(id string) (*image.Image, error) {
	return nil, fmt.Errorf("unsupported platform")
}

func (r *resolver) Extract(id string, l string, p string) error {
	return fmt.Errorf("unsupported platform")
}
