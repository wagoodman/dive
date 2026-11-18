package command

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wagoodman/dive/internal/source"
)

func TestBuildEngine(t *testing.T) {
	tests := map[string]string{
		source.Docker:            source.Docker,
		source.Podman:            source.Podman,
		source.ContainersStorage: source.Podman,
		source.Auto:              source.Docker,
		source.DockerArchive:     source.Docker,
		"docker-tar":             source.Docker,
		source.OCIArchive:        source.Docker,
		source.OCIDir:            source.Docker,
		source.Registry:          source.Docker,
		source.Containerd:        source.Docker,
		"":                       source.Docker,
	}
	for src, want := range tests {
		assert.Equal(t, want, buildEngine(src), src)
	}
}
