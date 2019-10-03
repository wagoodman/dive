package docker

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
)

type config struct {
	History []historyEntry `json:"history"`
	RootFs  rootFs         `json:"rootfs"`
}

type rootFs struct {
	Type    string   `json:"type"`
	DiffIds []string `json:"diff_ids"`
}

type historyEntry struct {
	ID         string
	Size       uint64
	Created    string `json:"created"`
	Author     string `json:"author"`
	CreatedBy  string `json:"created_by"`
	EmptyLayer bool   `json:"empty_layer"`
}

func newConfig(configBytes []byte) config {
	var imageConfig config
	err := json.Unmarshal(configBytes, &imageConfig)
	if err != nil {
		logrus.Panic(err)
	}

	layerIdx := 0
	for idx := range imageConfig.History {
		if imageConfig.History[idx].EmptyLayer {
			imageConfig.History[idx].ID = "<missing>"
		} else {
			imageConfig.History[idx].ID = imageConfig.RootFs.DiffIds[layerIdx]
			layerIdx++
		}
	}

	return imageConfig
}
