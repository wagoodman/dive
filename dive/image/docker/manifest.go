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
	manifests := newManifests(manifestBytes)
	if len(manifests) == 0 {
		panic(fmt.Errorf("empty manifest"))
	}
	return manifests[0]
}

func newManifests(manifestBytes []byte) []manifest {
	var manifests []manifest
	err := json.Unmarshal(manifestBytes, &manifests)
	if err == nil && len(manifests) > 0 {
		return manifests
	}

	var single manifest
	if singleErr := json.Unmarshal(manifestBytes, &single); singleErr == nil {
		return []manifest{single}
	}

	if err != nil {
		panic(fmt.Errorf("failed to unmarshal manifest: %w", err))
	}
	return nil
}
