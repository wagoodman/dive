package source

import (
	"context"
	"testing"

	"github.com/anchore/stereoscope/pkg/file"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// multiArchIndex is a linux/amd64 + linux/arm64 index with the "unknown/unknown" attestation manifest buildx adds
func multiArchIndex(t *testing.T) v1.ImageIndex {
	t.Helper()
	idx := mutate.IndexMediaType(empty.Index, types.OCIImageIndex)
	for _, arch := range []string{"amd64", "arm64"} {
		idx = mutate.AppendManifests(idx, mutate.IndexAddendum{
			Add:        archImage(t, arch),
			Descriptor: v1.Descriptor{Platform: &v1.Platform{OS: "linux", Architecture: arch}},
		})
	}
	return mutate.AppendManifests(idx, mutate.IndexAddendum{
		Add: archImage(t, "attestation"),
		Descriptor: v1.Descriptor{
			Platform:    &v1.Platform{OS: "unknown", Architecture: "unknown"},
			Annotations: map[string]string{"vnd.docker.reference.type": "attestation-manifest"},
		},
	})
}

// regression for #442: --platform picks one image out of a multi-arch index, whatever the host arch is
func TestLoad_platform(t *testing.T) {
	idx := multiArchIndex(t)

	dir := t.TempDir()
	_, err := layout.Write(dir, idx)
	require.NoError(t, err)

	host := startRegistry(t)
	ref := host + "/dive/multiarch:1"
	r, err := name.ParseReference(ref)
	require.NoError(t, err)
	require.NoError(t, remote.WriteIndex(r, idx))

	for src, input := range map[string]string{OCIDir: "oci-dir:" + dir, Registry: "registry://" + ref} {
		t.Run(src, func(t *testing.T) {
			for arch, other := range map[string]string{"amd64": "arm64", "arm64": "amd64"} {
				t.Run(arch, func(t *testing.T) {
					img, err := Load(context.Background(), input, Config{Platform: "linux/" + arch})
					require.NoError(t, err)
					t.Cleanup(func() { require.NoError(t, img.Cleanup()) })

					assert.Equal(t, arch, img.Metadata.Architecture)
					tree := img.SquashedTree()
					assert.True(t, tree.HasPath(file.Path("/"+arch)), "expected the %s image's files", arch)
					assert.False(t, tree.HasPath(file.Path("/"+other)), "got the %s image's files", other)
				})
			}

			// a platform the index doesn't have fails instead of quietly loading another one
			_, err := Load(context.Background(), input, Config{Platform: "linux/s390x"})
			require.Error(t, err)
		})
	}
}
