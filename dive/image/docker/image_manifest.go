package docker

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
)

type imageManifest struct {
	ConfigPath    string   `json:"Config"`
	RepoTags      []string `json:"RepoTags"`
	LayerTarPaths []string `json:"Layers"`
}

func newDockerImageManifest(manifestBytes []byte) imageManifest {
	var manifest []imageManifest
	err := json.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		logrus.Panic(err)
	}
	return manifest[0]
}
