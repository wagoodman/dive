package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/viewmodels"
	"go.uber.org/zap"
)

type LayersViewModel interface {
	SetLayerIndex(int) bool
	GetPrintableLayers() []fmt.Stringer
	SwitchMode()
	GetMode() viewmodels.LayerCompareMode
}

type LayerList struct {
	*tview.Box
	bufferIndexLowerBound int
	cmpIndex              int
	changed               LayerListHandler
	inputHandler          func(event *tcell.EventKey, setFocus func(p tview.Primitive))

	bindingArray []KeyBindingDisplay

	LayersViewModel
}

type LayerListHandler func(index int, shortcut rune)

func NewLayerList(model LayersViewModel) *LayerList {
	return &LayerList{
		Box:             tview.NewBox(),
		cmpIndex:        0,
		LayersViewModel: model,
		inputHandler:    nil,
	}
}

func (ll *LayerList) Setup(config KeyBindingConfig) *LayerList {
	bindingSettings := map[string]keyAction{
		"keybinding.page-up":   func() bool { return ll.pageUp() },
		"keybinding.page-down": func() bool { return ll.pageDown() },
		"keybinding.compare-all": func() bool {
			if ll.GetMode() == viewmodels.CompareSingleLayer {
				ll.SwitchMode()
				return true
			}
			return false
		},
		"keybinding.compare-layer": func() bool {
			if ll.GetMode() == viewmodels.CompareAllLayers {
				ll.SwitchMode()
				return true
			}
			return false
		},
	}

	actionArray := []keyAction{}

	for keybinding, action := range bindingSettings {
		binding, err := config.GetKeyBinding(keybinding)
		if err != nil {
			panic(fmt.Errorf("setup error during %s: %w", keybinding, err))
			// TODO handle this error
			//return nil
		}
		ll.bindingArray = append(ll.bindingArray, KeyBindingDisplay{KeyBinding: &binding, Selected: false})
		actionArray = append(actionArray, action)
	}

	ll.inputHandler = func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp, tcell.KeyLeft:
			if ll.SetLayerIndex(ll.cmpIndex - 1) {
				ll.keyUp()
				//ll.cmpIndex--
				//logrus.Debugf("KeyUp pressed, index: %d", ll.cmpIndex)
			}
		case tcell.KeyDown, tcell.KeyRight:
			if ll.SetLayerIndex(ll.cmpIndex + 1) {
				ll.keyDown()
				//ll.cmpIndex++
				//logrus.Debugf("KeyUp pressed, index: %d", ll.cmpIndex)

			}
		}
		for idx, binding := range ll.bindingArray {
			if binding.Match(event) {
				actionArray[idx]()
			}
		}
	}
	return ll
}

func (ll *LayerList) GetKeyBindings() []KeyBindingDisplay {
	return ll.bindingArray
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
		printWidth := intMin(len(line),width)
		format.PrintLine(screen, line, x, y+yIndex, printWidth, tview.AlignLeft,  tcell.StyleDefault)
		

	}
}

func (ll *LayerList) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return ll.WrapInputHandler(ll.inputHandler)
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

	logrus.Debugln("keyUp in layers")
	logrus.Debugf("  cmpIndex: %d", ll.cmpIndex)
	logrus.Debugf("  bufferIndexLowerBound: %d", ll.bufferIndexLowerBound)
	return true
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
	logrus.Debugln("keyDown in layers")
	logrus.Debugf("  cmpIndex: %d", ll.cmpIndex)
	logrus.Debugf("  bufferIndexLowerBound: %d", ll.bufferIndexLowerBound)

	return true
}

func (ll *LayerList) pageUp() bool {
	zap.S().Info("layer page up call")
	_, _, _, height := ll.Box.GetInnerRect()

	ll.cmpIndex = intMax(0, ll.cmpIndex-height)
	if ll.cmpIndex < ll.bufferIndexLowerBound {
		ll.bufferIndexLowerBound = ll.cmpIndex
	}

	return ll.SetLayerIndex(ll.cmpIndex)
}

func (ll *LayerList) pageDown() bool {
	zap.S().Info("layer page down call")
	// two parts of this are moving both the currently selected item & the window as a whole

	_, _, _, height := ll.Box.GetInnerRect()
	upperBoundIndex := len(ll.GetPrintableLayers()) - 1
	ll.cmpIndex = intMin(ll.cmpIndex+height, upperBoundIndex)
	if ll.cmpIndex >= ll.bufferIndexLowerBound+height {
		ll.bufferIndexLowerBound = intMin(ll.cmpIndex, upperBoundIndex-height+1)
	}

	return ll.SetLayerIndex(ll.cmpIndex)
}
