package source

import (
	"context"
	"testing"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// archImage is a one layer OCI image whose config says linux/<arch> and whose only file, /<arch>, names it, so a
// test can tell which image of an index was loaded from the files alone.
func archImage(t *testing.T, arch string) v1.Image {
	t.Helper()
	layer, err := crane.Layer(map[string][]byte{arch: []byte(arch)})
	require.NoError(t, err)
	base := mutate.ConfigMediaType(mutate.MediaType(empty.Image, types.OCIManifestSchema1), types.OCIConfigJSON)
	img, err := mutate.Append(base, mutate.Addendum{Layer: layer, MediaType: types.OCILayer})
	require.NoError(t, err)
	cf, err := img.ConfigFile()
	require.NoError(t, err)
	cf = cf.DeepCopy()
	cf.OS, cf.Architecture = "linux", arch
	img, err = mutate.ConfigFile(img, cf)
	require.NoError(t, err)
	return img
}

// regression for #423: an OCI image layout dir (skopeo `oci:<dir>`, `crane pull --format=oci`) loads via `oci-dir:`
func TestLoad_ociDirLayout(t *testing.T) {
	dir := t.TempDir()
	p, err := layout.Write(dir, empty.Index)
	require.NoError(t, err)
	require.NoError(t, p.AppendImage(archImage(t, "amd64")))

	for _, scheme := range []string{"oci-dir:", "oci-dir://"} {
		t.Run(scheme, func(t *testing.T) {
			img, err := Load(context.Background(), scheme+dir, Config{})
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, img.Cleanup()) })

			require.Len(t, img.Layers, 1)
			assert.True(t, img.SquashedTree().HasPath("/amd64"))
		})
	}

	// same layout through --source / container-engine instead of a scheme
	img, err := Load(context.Background(), dir, Config{Source: OCIDir})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, img.Cleanup()) })
	assert.Len(t, img.Layers, 1)
}
