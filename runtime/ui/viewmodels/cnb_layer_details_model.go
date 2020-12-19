package viewmodels

import (
	"github.com/buildpacks/lifecycle"
	"github.com/wagoodman/dive/dive/image"
)

type CNBLayersViewModel struct {
	*LayersViewModel
	bomMapping map[string]lifecycle.BOMEntry
}

func NewCNBLayersViewModel(layers []*image.Layer, bomMapping map[string]lifecycle.BOMEntry) *CNBLayersViewModel {
	return &CNBLayersViewModel{
		LayersViewModel: NewLayersViewModel(layers),
		bomMapping: bomMapping,
	}
}

func (cvm *CNBLayersViewModel) GetBOMFromDigest(layerSha string) lifecycle.BOMEntry{
	result, ok := cvm.bomMapping[layerSha]
	if !ok {
		return lifecycle.BOMEntry{}
	}
	return result
}
