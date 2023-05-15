//go:build !linux
// +build !linux

package podman

import (
	"fmt"
	"github.com/wagoodman/dive/dive/image"
)

type resolver struct{}

func NewResolverFromEngine() *resolver {
	return &resolver{}
}

func (r *resolver) Build(args []string) (*image.Image, error) {
	return nil, fmt.Errorf("unsupported platform")
}

func (r *resolver) Fetch(id string) (*image.Image, error) {
	return nil, fmt.Errorf("unsupported platform")
}
