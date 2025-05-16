package docker

import (
	"encoding/json"
	"fmt"
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
