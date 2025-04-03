package docker

import (
	"encoding/json"
	"path/filepath"

	digest "github.com/opencontainers/go-digest"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
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
		logrus.Panic(err)
	}
	return manifest[0]
}

func newOCIManifest(manifestBytes []byte) oci.Manifest {
	ociManifest := oci.Manifest{}
	err := json.Unmarshal(manifestBytes, &ociManifest)
	if err != nil {
		logrus.Panic(err)
	}
	if ociManifest.MediaType != oci.MediaTypeImageManifest {
		logrus.Panicf("mediaType mismatch: expected '%s', found '%s'", oci.MediaTypeImageManifest, ociManifest.MediaType)
	}
	return ociManifest
}

func extractManifest(indexBytes []byte) oci.Descriptor {
	ociIndex := oci.Index{}
	err := json.Unmarshal(indexBytes, &ociIndex)
	if err != nil {
		logrus.Panic(err)
	}
	manifests := ociIndex.Manifests
	if len(manifests) == 0 {
		logrus.Panic("No manifest found")
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
