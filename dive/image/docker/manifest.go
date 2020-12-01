package docker

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
)

type Manifest struct {
	ConfigPath    string   `json:"Config"`
	RepoTags      []string `json:"RepoTags"`
	LayerTarPaths []string `json:"Layers"`
}

func newManifest(manifestBytes []byte) Manifest {
	var manifest []Manifest
	err := json.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		logrus.Panic(err)
	}
	return manifest[0]
}
