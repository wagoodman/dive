package docker

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wagoodman/dive/dive/image"
)

func createTestTarLayer(t *testing.T, filename string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	err := tw.WriteHeader(&tar.Header{
		Name: filename,
		Mode: 0644,
		Size: int64(len(content)),
	})
	require.NoError(t, err)
	_, err = tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	return buf.Bytes()
}

func gzipBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(data)
	require.NoError(t, err)
	require.NoError(t, gw.Close())
	return buf.Bytes()
}

func zstdBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw, err := zstd.NewWriter(&buf)
	require.NoError(t, err)
	_, err = zw.Write(data)
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

func buildTestArchive(t *testing.T, entries map[string][]byte, links map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	for name, content := range entries {
		err := tw.WriteHeader(&tar.Header{
			Name:     name,
			Mode:     0644,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		})
		require.NoError(t, err)
		_, err = tw.Write(content)
		require.NoError(t, err)
	}

	for linkName, targetName := range links {
		err := tw.WriteHeader(&tar.Header{
			Name:     linkName,
			Linkname: targetName,
			Mode:     0644,
			Typeflag: tar.TypeLink,
		})
		require.NoError(t, err)
	}

	require.NoError(t, tw.Close())
	return buf.Bytes()
}

func TestLayerResolutionBlobsSha256(t *testing.T) {
	layerHash1 := "efa9dd3c982f5c3afad3593d56afbc731dd6f1b8661661f3775f0f4a82a436bd"
	layerHash2 := "9028f16e6cba11223344556677889900aabbccddeeff00112233445566778899"
	configHash := "c6e552ba03945bd8a68c68d7c65cbd15b553ca4a5caf17731718cb41994c4283"

	layer1Tar := createTestTarLayer(t, "app/main.go", []byte("package main\nfunc main() {}"))
	layer1Gz := gzipBytes(t, layer1Tar)

	layer2Tar := createTestTarLayer(t, "app/config.json", []byte(`{"env":"prod"}`))
	layer2Zstd := zstdBytes(t, layer2Tar)

	cfg := config{
		History: []historyEntry{
			{CreatedBy: "RUN step 1"},
			{CreatedBy: "RUN step 2"},
		},
		RootFs: rootFs{
			Type: "layers",
			DiffIds: []string{
				"sha256:" + layerHash1,
				"sha256:" + layerHash2,
			},
		},
	}
	cfgBytes, err := json.Marshal(cfg)
	require.NoError(t, err)

	manifestObj := manifest{
		ConfigPath: "blobs/sha256/" + configHash,
		RepoTags:   []string{"test:latest"},
		LayerTarPaths: []string{
			"blobs/sha256/" + layerHash1,
			"blobs/sha256/" + layerHash2,
		},
	}
	manifestBytes, err := json.Marshal([]manifest{manifestObj})
	require.NoError(t, err)

	archiveData := buildTestArchive(t, map[string][]byte{
		"manifest.json":                     manifestBytes,
		"blobs/sha256/" + configHash:        cfgBytes,
		"blobs/sha256/" + layerHash1:        layer1Gz,
		"blobs/sha256/" + layerHash2:        layer2Zstd,
	}, nil)

	archive, err := NewImageArchive(io.NopCloser(bytes.NewReader(archiveData)))
	require.NoError(t, err)

	img, err := archive.ToImage("test:latest")
	require.NoError(t, err)
	require.Len(t, img.Layers, 2)
	assert.Equal(t, layerHash1, img.Layers[0].Id)
	assert.Equal(t, layerHash2, img.Layers[1].Id)

	analysis, err := image.Analyze(context.Background(), img)
	require.NoError(t, err)
	assert.NotNil(t, analysis)
}

func TestLayerResolutionLegacyAndOciCrossFormat(t *testing.T) {
	layerHash := "1111222233334444555566667777888899990000aaaabbbbccccddddeeeeffff"
	configHash := "9999888877776666555544443333222211110000ffffeeeeddddccccbbbbaaaa"

	layerTar := createTestTarLayer(t, "data.txt", []byte("cross format test data"))

	cfg := config{
		History: []historyEntry{{CreatedBy: "COPY data.txt"}},
		RootFs:  rootFs{Type: "layers", DiffIds: []string{"sha256:" + layerHash}},
	}
	cfgBytes, err := json.Marshal(cfg)
	require.NoError(t, err)

	t.Run("manifest specifies blobs/sha256 but archive stores legacy id/layer.tar", func(t *testing.T) {
		manifestObj := manifest{
			ConfigPath:    configHash + ".json",
			RepoTags:      []string{"legacy-tar:latest"},
			LayerTarPaths: []string{"blobs/sha256/" + layerHash},
		}
		manifestBytes, err := json.Marshal([]manifest{manifestObj})
		require.NoError(t, err)

		archiveData := buildTestArchive(t, map[string][]byte{
			"manifest.json":            manifestBytes,
			configHash + ".json":       cfgBytes,
			layerHash + "/layer.tar":   layerTar,
		}, nil)

		archive, err := NewImageArchive(io.NopCloser(bytes.NewReader(archiveData)))
		require.NoError(t, err)

		img, err := archive.ToImage("legacy-tar:latest")
		require.NoError(t, err)
		require.Len(t, img.Layers, 1)
		assert.Equal(t, layerHash, img.Layers[0].Id)
	})

	t.Run("manifest specifies legacy id/layer.tar but archive stores blobs/sha256", func(t *testing.T) {
		manifestObj := manifest{
			ConfigPath:    "blobs/sha256/" + configHash,
			RepoTags:      []string{"oci-store:latest"},
			LayerTarPaths: []string{layerHash + "/layer.tar"},
		}
		manifestBytes, err := json.Marshal([]manifest{manifestObj})
		require.NoError(t, err)

		archiveData := buildTestArchive(t, map[string][]byte{
			"manifest.json":              manifestBytes,
			"blobs/sha256/" + configHash: cfgBytes,
			"blobs/sha256/" + layerHash:  layerTar,
		}, nil)

		archive, err := NewImageArchive(io.NopCloser(bytes.NewReader(archiveData)))
		require.NoError(t, err)

		img, err := archive.ToImage("oci-store:latest")
		require.NoError(t, err)
		require.Len(t, img.Layers, 1)
		assert.Equal(t, layerHash, img.Layers[0].Id)
	})

	t.Run("manifest specifies dot-slash relative blobs path", func(t *testing.T) {
		manifestObj := manifest{
			ConfigPath:    "./blobs/sha256/" + configHash,
			RepoTags:      []string{"relative-path:latest"},
			LayerTarPaths: []string{"./blobs/sha256/" + layerHash},
		}
		manifestBytes, err := json.Marshal([]manifest{manifestObj})
		require.NoError(t, err)

		archiveData := buildTestArchive(t, map[string][]byte{
			"./manifest.json":            manifestBytes,
			"./blobs/sha256/" + configHash: cfgBytes,
			"./blobs/sha256/" + layerHash:  layerTar,
		}, nil)

		archive, err := NewImageArchive(io.NopCloser(bytes.NewReader(archiveData)))
		require.NoError(t, err)

		img, err := archive.ToImage("relative-path:latest")
		require.NoError(t, err)
		require.Len(t, img.Layers, 1)
		assert.Equal(t, layerHash, img.Layers[0].Id)
	})
}

func TestLayerResolutionHardlinks(t *testing.T) {
	layerHash1 := "aaaa1111bbbb2222cccc3333dddd4444eeee5555ffff6666aaaa1111bbbb2222"
	layerHash2 := "bbbb2222cccc3333dddd4444eeee5555ffff6666aaaa1111bbbb2222cccc3333"
	configHash := "cccc3333dddd4444eeee5555ffff6666aaaa1111bbbb2222cccc3333dddd4444"

	layerTar := createTestTarLayer(t, "shared/lib.so", []byte("binary data"))

	cfg := config{
		History: []historyEntry{
			{CreatedBy: "base stage"},
			{CreatedBy: "derived stage duplicate"},
		},
		RootFs: rootFs{
			Type: "layers",
			DiffIds: []string{
				"sha256:" + layerHash1,
				"sha256:" + layerHash2,
			},
		},
	}
	cfgBytes, err := json.Marshal(cfg)
	require.NoError(t, err)

	manifestObj := manifest{
		ConfigPath: "blobs/sha256/" + configHash,
		RepoTags:   []string{"hardlink-test:latest"},
		LayerTarPaths: []string{
			"blobs/sha256/" + layerHash1,
			"blobs/sha256/" + layerHash2,
		},
	}
	manifestBytes, err := json.Marshal([]manifest{manifestObj})
	require.NoError(t, err)

	archiveData := buildTestArchive(t, map[string][]byte{
		"manifest.json":              manifestBytes,
		"blobs/sha256/" + configHash: cfgBytes,
		"blobs/sha256/" + layerHash1: layerTar,
	}, map[string]string{
		"blobs/sha256/" + layerHash2: "blobs/sha256/" + layerHash1,
	})

	archive, err := NewImageArchive(io.NopCloser(bytes.NewReader(archiveData)))
	require.NoError(t, err)

	img, err := archive.ToImage("hardlink-test:latest")
	require.NoError(t, err)
	require.Len(t, img.Layers, 2)
	assert.Equal(t, layerHash1, img.Layers[0].Id)
	assert.Equal(t, layerHash2, img.Layers[1].Id)
}

func TestLayerResolutionUncompressedAndEmptyBlob(t *testing.T) {
	layerHash1 := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	layerHashEmpty := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	configHash := "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"

	layerTar := createTestTarLayer(t, "file.txt", []byte("uncompressed tar layer"))

	cfg := config{
		History: []historyEntry{
			{CreatedBy: "layer 1"},
			{CreatedBy: "empty layer"},
		},
		RootFs: rootFs{
			Type: "layers",
			DiffIds: []string{
				"sha256:" + layerHash1,
				"sha256:" + layerHashEmpty,
			},
		},
	}
	cfgBytes, err := json.Marshal(cfg)
	require.NoError(t, err)

	manifestObj := manifest{
		ConfigPath: "blobs/sha256/" + configHash,
		RepoTags:   []string{"empty-blob:latest"},
		LayerTarPaths: []string{
			"blobs/sha256/" + layerHash1,
			"blobs/sha256/" + layerHashEmpty,
		},
	}
	manifestBytes, err := json.Marshal([]manifest{manifestObj})
	require.NoError(t, err)

	archiveData := buildTestArchive(t, map[string][]byte{
		"manifest.json":                 manifestBytes,
		"blobs/sha256/" + configHash:    cfgBytes,
		"blobs/sha256/" + layerHash1:    layerTar,
		"blobs/sha256/" + layerHashEmpty: []byte{},
	}, nil)

	archive, err := NewImageArchive(io.NopCloser(bytes.NewReader(archiveData)))
	require.NoError(t, err)

	img, err := archive.ToImage("empty-blob:latest")
	require.NoError(t, err)
	require.Len(t, img.Layers, 2)
}

func TestOciDataTarArchives(t *testing.T) {
	t.Run("test-oci-docker-image.tar", func(t *testing.T) {
		result := TestAnalysisFromArchive(t, "../../../.data/test-oci-docker-image.tar")
		assert.NotNil(t, result)
		assert.True(t, result.SizeBytes > 0)
	})

	t.Run("test-docker-image.tar", func(t *testing.T) {
		result := TestAnalysisFromArchive(t, "../../../.data/test-docker-image.tar")
		assert.NotNil(t, result)
		assert.True(t, result.SizeBytes > 0)
	})
}
