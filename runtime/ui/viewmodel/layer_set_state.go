package viewmodel

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/format"
)

type LayerSetState struct {
	LayerIndex        int
	Layers            []*image.Layer
	CompareMode       LayerCompareMode
	CompareStartIndex int

	constrainedRealEstate bool
	viewStartIndex int
	viewHeight     int

	Buffer bytes.Buffer
}

func NewLayerSetState(layers []*image.Layer, compareMode LayerCompareMode) *LayerSetState {
	return &LayerSetState{
		Layers:      layers,
		CompareMode: compareMode,
		LayerIndex:     0,
		viewStartIndex: 0,
	}
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (vm *LayerSetState) Setup(lowerBound, height int) {
	vm.viewStartIndex = lowerBound
	vm.viewHeight = height
}

// height returns the current height and considers the header
func (vm *LayerSetState) height() int {
	return vm.viewHeight - 1
}

// IsVisible indicates if the layer view pane is currently initialized
func (vm *LayerSetState) IsVisible() bool {
	return vm != nil
}

// ResetCursor moves the cursor back to the top of the buffer and translates to the top of the buffer.
func (vm *LayerSetState) ResetCursor() {
	vm.LayerIndex = 0
	vm.viewStartIndex = 0
}


// PageUp moves to previous page putting the cursor on top
func (vm *LayerSetState) PageUp() bool {
	prevPageEndIndex := vm.viewStartIndex
	prevPageStartIndex := vm.viewStartIndex - vm.viewHeight + 1

	if prevPageStartIndex < 0 {
		prevPageStartIndex = 0
		vm.LayerIndex = 0
		prevPageEndIndex = vm.viewHeight
		if prevPageEndIndex >= len(vm.Layers) {
			return false
		}
	}

	vm.viewStartIndex = prevPageStartIndex

	if vm.LayerIndex >= prevPageEndIndex {
		vm.LayerIndex = prevPageEndIndex
	}
	return true
}

// PageDown moves to next page putting the cursor on top
func (vm *LayerSetState) PageDown() bool {
	nextPageStartIndex := vm.viewStartIndex + vm.viewHeight - 1
	nextPageEndIndex := nextPageStartIndex + vm.viewHeight

	if nextPageEndIndex > len(vm.Layers) {
		nextPageEndIndex = len(vm.Layers) - 1
		vm.LayerIndex = nextPageEndIndex
		nextPageStartIndex = nextPageEndIndex - vm.viewHeight + 1
		if (nextPageStartIndex < 0) {
			return false
		}
	}

	vm.viewStartIndex = nextPageStartIndex

	if vm.LayerIndex < nextPageStartIndex {
		vm.LayerIndex = nextPageStartIndex
	}
	
	return true
}

// doCursorUp performs the internal view's adjustments on cursor up. Note: this is independent of the gocui buffer.
func (vm *LayerSetState) CursorUp() bool {
	if vm.LayerIndex <= 0 {
		return false
	}
	vm.LayerIndex--
	if vm.LayerIndex < vm.viewStartIndex {
		vm.viewStartIndex--
	}
	return true
}

// doCursorDown performs the internal view's adjustments on cursor down. Note: this is independent of the gocui buffer.
func (vm *LayerSetState) CursorDown() bool {
	if vm.LayerIndex >= len(vm.Layers) - 1 {
		return false
	}
	vm.LayerIndex++
	if vm.LayerIndex >= vm.viewStartIndex + vm.viewHeight {
		vm.viewStartIndex++
	}
	return true
}

// renderCompareBar returns the formatted string for the given layer.
func (vm *LayerSetState) renderCompareBar(layerIdx int) string {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop := vm.GetCompareIndexes()
	result := "  "

	if layerIdx >= bottomTreeStart && layerIdx <= bottomTreeStop {
		result = format.CompareBottom("  ")
	}
	if layerIdx >= topTreeStart && layerIdx <= topTreeStop {
		result = format.CompareTop("  ")
	}

	return result
}

// Update refreshes the state objects for future rendering
func (vm *LayerSetState) Update(isConstrainedRealEstate bool) error {
	vm.constrainedRealEstate = isConstrainedRealEstate
	return nil
}

// Render flushes the state objects to the screen. The layers pane reports:
// 1. the layers of the image surrounding the currently selected layer
// 2. the current selected layer
func (vm *LayerSetState) Render() error {
	logrus.Tracef("viewmodel.LayerSetState.Render() %s", vm.Layers[vm.LayerIndex].Id)

	// write contents of pane
	vm.Buffer.Reset()
	for idx, layer := range vm.Layers {
		if idx < vm.viewStartIndex {
			continue
		}
		if idx > vm.viewStartIndex + vm.viewHeight {
			break
		}
		var layerStr string
		if vm.constrainedRealEstate {
			layerStr = fmt.Sprintf("%-4d", layer.Index)
		} else {
			layerStr = layer.String()
		}

		compareBar := vm.renderCompareBar(idx)

		err := error(nil)
		if idx == vm.LayerIndex {
			_, err = fmt.Fprintln(&vm.Buffer, compareBar+" "+format.Selected(layerStr))
		} else {
			_, err = fmt.Fprintln(&vm.Buffer, compareBar+" "+layerStr)
		}

		if err != nil {
			logrus.Debug("unable to write to buffer: ", err)
			return err
		}
	}
	return nil
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
