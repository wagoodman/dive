package docker

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	digest "github.com/opencontainers/go-digest"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

type manifest struct {
	ConfigPath    string   `json:"Config"`
	RepoTags      []string `json:"RepoTags"`
	LayerTarPaths []string `json:"Layers"`
}

func newManifest(manifestBytes []byte) manifest {
	var manifest []manifest
	err := json.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal manifest: %w", err))
	}
	return manifest[0]
}

func newOCIManifest(manifestBytes []byte) oci.Manifest {
	ociManifest := oci.Manifest{}
	err := json.Unmarshal(manifestBytes, &ociManifest)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal manifest: %w", err))
	}
	if ociManifest.MediaType != oci.MediaTypeImageManifest {
		panic(fmt.Errorf("mediaType mismatch: expected '%s', found '%s'", oci.MediaTypeImageManifest, ociManifest.MediaType))
	}
	return ociManifest
}

func extractManifest(indexBytes []byte) oci.Descriptor {
	ociIndex := oci.Index{}
	err := json.Unmarshal(indexBytes, &ociIndex)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal index: %w", err))
	}
	manifests := ociIndex.Manifests
	if len(manifests) == 0 {
		panic(fmt.Errorf("No manifest found"))
	}
	return manifests[0]
}

func digestPath(d digest.Digest) string {
	return filepath.Join(oci.ImageBlobsDir, d.Algorithm().String(), d.Encoded())
}

func layerPaths(ociManifest oci.Manifest) []string {
	var layerPaths []string
	for _, path := range ociManifest.Layers {
		layerPaths = append(layerPaths, digestPath(path.Digest))
	}
	return layerPaths
}
