package skopeo

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/image/docker"
	"strings"
)

type manifestV2 struct {
	SchemaVersion int     `json:"schemaVersion"`
	MediaType     string  `json:"mediaType"`
	Config        config  `json:"config"`
	Layers        []layer `json:"layers"`
}

type config struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

type layer struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

func newManifest(manifestBytes []byte) docker.Manifest {
	var manifest manifestV2
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		logrus.Panic(err)
	}
	var layers []string
	for _, l := range manifest.Layers {
		layers = append(layers, strings.TrimPrefix(l.Digest, "sha256:"))
	}
	return docker.Manifest{
		ConfigPath:    strings.TrimPrefix(manifest.Config.Digest, "sha256:"),
		LayerTarPaths: layers,
	}
}
