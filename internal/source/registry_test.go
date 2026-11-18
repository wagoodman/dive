package source

import (
	"context"
	"io"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startRegistry serves an in-memory registry on loopback and returns its host:port. DOCKER_CONFIG points at an empty
// dir so no real credentials or credential helpers get involved.
func startRegistry(t *testing.T) string {
	t.Helper()
	t.Setenv("DOCKER_CONFIG", t.TempDir())
	srv := httptest.NewServer(registry.New(registry.Logger(log.New(io.Discard, "", 0))))
	t.Cleanup(srv.Close)
	return strings.TrimPrefix(srv.URL, "http://")
}

// regression for #406 and #130: an image is read straight from a registry with go-containerregistry, no docker
// daemon (or skopeo) involved
func TestLoad_registry(t *testing.T) {
	host := startRegistry(t)
	ref := host + "/dive/app:1"
	r, err := name.ParseReference(ref)
	require.NoError(t, err)
	require.NoError(t, remote.Write(r, archImage(t, "amd64")))

	tests := []struct {
		name  string
		input string
		cfg   Config
	}{
		{name: "registry scheme", input: "registry://" + ref},
		{name: "oci-registry scheme", input: "oci-registry://" + ref},
		{name: "registry source", input: ref, cfg: Config{Source: Registry}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := Load(context.Background(), tt.input, tt.cfg)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, img.Cleanup()) })

			require.Len(t, img.Layers, 1)
			assert.True(t, img.SquashedTree().HasPath("/amd64"))
		})
	}

	// a tag the registry doesn't have is an error, not a fallback to some other source
	_, err = Load(context.Background(), "registry://"+host+"/dive/app:missing", Config{})
	require.Error(t, err)
}
