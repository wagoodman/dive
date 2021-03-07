package components

import (
	"fmt"

	"github.com/wagoodman/dive/runtime/config"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wagoodman/dive/internal/log"
	"github.com/wagoodman/dive/runtime/ui/components/helpers"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/viewmodels"
)

type LayersViewModel interface {
	SetLayerIndex(int) bool
	GetPrintableLayers() []fmt.Stringer
	SwitchLayerMode() error
	GetMode() viewmodels.LayerCompareMode
}

type LayerList struct {
	*tview.Box
	bufferIndexLowerBound int
	cmpIndex              int
	changed               LayerListHandler

	keyInputHandler *helpers.KeyInputHandler

	LayersViewModel
}

type LayerListHandler func(index int, shortcut rune)

func NewLayerList(model LayersViewModel) *LayerList {
	return &LayerList{
		Box:             tview.NewBox(),
		cmpIndex:        0,
		LayersViewModel: model,
		keyInputHandler: helpers.NewKeyInputHandler(),
	}
}

type layerListViewOption func(ll *LayerList)

var alwaysFalse = func() bool { return false }
var alwaysTrue = func() bool { return true }

func upLayerListBindingOption() layerListViewOption {
	k := helpers.NewKeyBinding("Cursor Up", tcell.NewEventKey(tcell.KeyUp, rune(0), tcell.ModNone))
	return func(ll *LayerList) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   alwaysFalse,
			Hide:       alwaysTrue,
		}
		ll.keyInputHandler.AddBinding(displayBinding, func() { ll.keyUp() })
	}
}

func downLayerListBindingOption() layerListViewOption {
	k := helpers.NewKeyBinding("Cursor Down", tcell.NewEventKey(tcell.KeyDown, rune(0), tcell.ModNone))

	return func(ll *LayerList) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   alwaysFalse,
			Hide:       alwaysTrue,
		}
		ll.keyInputHandler.AddBinding(displayBinding, func() { ll.keyDown() })
	}
}

func pageUpLayerListBindingOption(bindingValue string) layerListViewOption {
	k := helpers.NewKeyBinding("Pg Up", helpers.DecodeBinding(bindingValue))
	return func(ll *LayerList) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   alwaysFalse,
			Hide:       alwaysTrue,
		}
		ll.keyInputHandler.AddBinding(displayBinding, func() { ll.pageUp() })
	}
}

func pageDownLayerListBindingOption(bindingValue string) layerListViewOption {
	k := helpers.NewKeyBinding("Pg Down", helpers.DecodeBinding(bindingValue))
	return func(ll *LayerList) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   alwaysFalse,
			Hide:       alwaysTrue,
		}
		ll.keyInputHandler.AddBinding(displayBinding, func() { ll.pageDown() })
	}
}

func compareAllLayerListBindingOption(bindingValue string) layerListViewOption {
	k := helpers.NewKeyBinding("Aggregate Changes", helpers.DecodeBinding(bindingValue))
	return func(ll *LayerList) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   func() bool { return ll.GetMode() == viewmodels.CompareAllLayers },
			Hide:       alwaysFalse,
		}
		ll.keyInputHandler.AddBinding(displayBinding, func() {
			// TODO: swap out switch for set
			if err := ll.SwitchLayerMode(); err != nil {
				log.WithFields("error", err).Error("CompareAllLayers failed")
			}
		})
	}
}

func compareSingleLayerListBindingOption(bindingValue string) layerListViewOption {
	k := helpers.NewKeyBinding("Layer Changes", helpers.DecodeBinding(bindingValue))
	return func(ll *LayerList) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   func() bool { return ll.GetMode() == viewmodels.CompareSingleLayer },
			Hide:       alwaysFalse,
		}
		ll.keyInputHandler.AddBinding(displayBinding, func() {
			// TODO: swap out switch for set
			if err := ll.SwitchLayerMode(); err != nil {
				log.WithFields("error", err).Error("CompareSingleLayer failed")
			}
		})
	}
}

func (ll *LayerList) AddBindingOptions(bindingOptions ...layerListViewOption) *LayerList {
	for _, option := range bindingOptions {
		option(ll)
	}

	return ll
}

func (ll *LayerList) Setup(cfg config.KeybindingConfig) *LayerList {
	ll.AddBindingOptions(
		upLayerListBindingOption(),
		downLayerListBindingOption(),
		pageUpLayerListBindingOption(cfg.PageUp),
		pageDownLayerListBindingOption(cfg.PageDown),
		compareSingleLayerListBindingOption(cfg.CompareLayer),
		compareAllLayerListBindingOption(cfg.CompareAll),
	)
	return ll
}

func (ll *LayerList) GetKeyBindings() []helpers.KeyBindingDisplay {
	return ll.keyInputHandler.Order
}

func (ll *LayerList) getBox() *tview.Box {
	return ll.Box
}

func (ll *LayerList) getDraw() drawFn {
	return ll.Draw
}

func (ll *LayerList) getInputWrapper() inputFn {
	return ll.InputHandler
}

func (ll *LayerList) Draw(screen tcell.Screen) {
	ll.Box.Draw(screen)
	x, y, width, height := ll.Box.GetInnerRect()
	compressedView := width < 25

	cmpString := "  "
	printableLayers := ll.GetPrintableLayers()
	for yIndex := 0; yIndex < height; yIndex++ {
		layerIndex := ll.bufferIndexLowerBound + yIndex
		if layerIndex >= len(printableLayers) {
			break
		}
		layer := printableLayers[layerIndex]
		cmpFormatter := format.Normal
		lineFormatter := format.Normal
		switch {
		case layerIndex == ll.cmpIndex:
			cmpFormatter = format.CompareTop
			lineFormatter = format.Selected
		case layerIndex > 0 && layerIndex < ll.cmpIndex && ll.GetMode() == viewmodels.CompareAllLayers:
			cmpFormatter = format.CompareTop
		case layerIndex < ll.cmpIndex:
			cmpFormatter = format.CompareBottom
		}
		line := fmt.Sprintf("%s %s", cmpFormatter(cmpString), lineFormatter(layer.String()))
		if compressedView {
			line = fmt.Sprintf("%s %s", cmpFormatter(cmpString), lineFormatter(fmt.Sprintf("%d", yIndex+1)))
		}
		printWidth := intMin(len(line), width)
		format.PrintLine(screen, line, x, y+yIndex, printWidth, tview.AlignLeft, tcell.StyleDefault)
	}
}

func (ll *LayerList) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return ll.keyInputHandler.Handle()
}

func (ll *LayerList) Focus(delegate func(p tview.Primitive)) {
	ll.Box.Focus(delegate)
}

func (ll *LayerList) HasFocus() bool {
	return ll.Box.HasFocus()
}

func (ll *LayerList) SetChangedFunc(handler LayerListHandler) *LayerList {
	ll.changed = handler
	return ll
}

func (ll *LayerList) keyUp() bool {
	if ll.cmpIndex <= 0 {
		return false
	}
	ll.cmpIndex--
	if ll.cmpIndex < ll.bufferIndexLowerBound {
		ll.bufferIndexLowerBound--
	}

	log.WithFields(
		"component", "LayerList",
		"cmpIndex", ll.cmpIndex,
		"bufferIndexLowerBound", ll.bufferIndexLowerBound,
	).Tracef("keyUp event")

	return ll.SetLayerIndex(ll.cmpIndex)
}

// TODO (simplify all page increments to rely an a single function)
func (ll *LayerList) keyDown() bool {
	_, _, _, height := ll.Box.GetInnerRect()

	visibleSize := len(ll.GetPrintableLayers())
	if ll.cmpIndex+1 >= visibleSize {
		return false
	}
	ll.cmpIndex++
	if ll.cmpIndex-ll.bufferIndexLowerBound >= height {
		ll.bufferIndexLowerBound++
	}

	log.WithFields(
		"component", "LayerList",
		"cmpIndex", ll.cmpIndex,
		"bufferIndexLowerBound", ll.bufferIndexLowerBound,
	).Tracef("keyDown event")

	return ll.SetLayerIndex(ll.cmpIndex)
}

func (ll *LayerList) pageUp() bool {
	log.WithFields(
		"component", "LayerList",
	).Tracef("pageUp event")

	_, _, _, height := ll.Box.GetInnerRect()

	ll.cmpIndex = intMax(0, ll.cmpIndex-height)
	if ll.cmpIndex < ll.bufferIndexLowerBound {
		ll.bufferIndexLowerBound = ll.cmpIndex
	}

	return ll.SetLayerIndex(ll.cmpIndex)
}

func (ll *LayerList) pageDown() bool {
	log.WithFields(
		"component", "LayerList",
	).Tracef("pageDown event")
	// two parts of this are moving both the currently selected item & the window as a whole

	_, _, _, height := ll.Box.GetInnerRect()
	upperBoundIndex := len(ll.GetPrintableLayers()) - 1
	ll.cmpIndex = intMin(ll.cmpIndex+height, upperBoundIndex)
	if ll.cmpIndex >= ll.bufferIndexLowerBound+height {
		ll.bufferIndexLowerBound = intMin(ll.cmpIndex, upperBoundIndex-height+1)
	}

	return ll.SetLayerIndex(ll.cmpIndex)
}
