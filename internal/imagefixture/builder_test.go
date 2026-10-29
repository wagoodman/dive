package imagefixture

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/require"
)

func TestFixturesAreDeterministic(t *testing.T) {
	for name, spec := range Fixtures() {
		var a, b bytes.Buffer
		require.NoError(t, spec.Write(&a), name)
		require.NoError(t, spec.Write(&b), name)
		require.Equal(t, a.Bytes(), b.Bytes(), name)
	}
}

func TestOCISaveDiffIDs(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Fixtures()["oci-save-zstd"].Write(&buf))
	files := map[string][]byte{}
	tr := tar.NewReader(&buf)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		files[hdr.Name], _ = io.ReadAll(tr)
	}

	var manifest []dockerManifest
	require.NoError(t, json.Unmarshal(files["manifest.json"], &manifest))
	var cfg struct {
		RootFS struct {
			DiffIDs []string `json:"diff_ids"`
		} `json:"rootfs"`
	}
	require.NoError(t, json.Unmarshal(files[manifest[0].Config], &cfg))
	require.Len(t, cfg.RootFS.DiffIDs, len(manifest[0].Layers))

	// diff ids are digests of the uncompressed layer tars
	for i, p := range manifest[0].Layers {
		dec, err := zstd.NewReader(bytes.NewReader(files[p]))
		require.NoError(t, err)
		raw, err := io.ReadAll(dec)
		require.NoError(t, err)
		sum := sha256.Sum256(raw)
		require.Equal(t, cfg.RootFS.DiffIDs[i], "sha256:"+hex.EncodeToString(sum[:]))
	}
}
