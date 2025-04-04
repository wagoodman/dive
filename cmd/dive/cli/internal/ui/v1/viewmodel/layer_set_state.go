package viewmodel

import "github.com/wagoodman/dive/dive/image"

type LayerSetState struct {
	LayerIndex        int
	Layers            []*image.Layer
	CompareMode       LayerCompareMode
	CompareStartIndex int
}

func NewLayerSetState(layers []*image.Layer, compareMode LayerCompareMode) *LayerSetState {
	return &LayerSetState{
		Layers:      layers,
		CompareMode: compareMode,
	}
}

// getCompareIndexes determines the layer boundaries to use for comparison (based on the current compare mode)
func (state *LayerSetState) GetCompareIndexes() (bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) {
	bottomTreeStart = state.CompareStartIndex
	topTreeStop = state.LayerIndex

	if state.LayerIndex == state.CompareStartIndex {
		bottomTreeStop = state.LayerIndex
		topTreeStart = state.LayerIndex
	} else if state.CompareMode == CompareSingleLayer {
		bottomTreeStop = state.LayerIndex - 1
		topTreeStart = state.LayerIndex
	} else {
		bottomTreeStop = state.CompareStartIndex
		topTreeStart = state.CompareStartIndex + 1
	}

	return bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop
}
