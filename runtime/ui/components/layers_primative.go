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

	keyInputHandler *KeyInputHandler

	LayersViewModel
}

type LayerListHandler func(index int, shortcut rune)

func NewLayerList(model LayersViewModel) *LayerList {
	return &LayerList{
		Box:             tview.NewBox(),
		cmpIndex:        0,
		LayersViewModel: model,
		keyInputHandler: NewKeyInputHandler(),
	}
}

type LayerListViewOption func(ll *LayerList)

func UpLayerListBindingOption(k KeyBindingDisplay) LayerListViewOption {
	return func(ll *LayerList) {
		ll.keyInputHandler.AddBinding(k, func() { ll.keyUp() })
	}
}

func DownLayerListBindingOption(k KeyBindingDisplay) LayerListViewOption {
	return func(ll *LayerList) {
		ll.keyInputHandler.AddBinding(k, func() { ll.keyDown() })
	}
}

func PageUpLayerListBindingOption(k KeyBindingDisplay) LayerListViewOption {
	return func(ll *LayerList) {
		ll.keyInputHandler.AddBinding(k, func() { ll.pageUp() })
	}
}

func PageDownLayerListBindingOption(k KeyBindingDisplay) LayerListViewOption {
	return func(ll *LayerList) {
		ll.keyInputHandler.AddBinding(k, func() { ll.pageDown() })
	}
}

func SwitchCompareLayerListBindingOption(k KeyBindingDisplay) LayerListViewOption {
	return func(ll *LayerList) { 
		ll.keyInputHandler.AddToggleBinding(k, func() { ll.SwitchMode() })
	}
}

func (ll *LayerList) AddBindingOptions(bindingOptions ...LayerListViewOption) *LayerList {
	for _, option := range bindingOptions {
		option(ll)
	}

	return ll
}

func (ll *LayerList) Setup(config KeyBindingConfig) *LayerList {

	ll.AddBindingOptions(
		UpLayerListBindingOption(NewKeyBindingDisplay(tcell.KeyUp, rune(0), tcell.ModNone, "", false, true)),
		UpLayerListBindingOption(NewKeyBindingDisplay(tcell.KeyLeft, rune(0), tcell.ModNone, "", false, true)),
		DownLayerListBindingOption(NewKeyBindingDisplay(tcell.KeyDown, rune(0), tcell.ModNone, "", false, true)),
		DownLayerListBindingOption(NewKeyBindingDisplay(tcell.KeyRight, rune(0), tcell.ModNone, "", false, true)),
	)

	bindingSettings := map[string]func(KeyBindingDisplay) LayerListViewOption{
		"keybinding.page-up":       PageUpLayerListBindingOption,
		"keybinding.page-down":     PageDownLayerListBindingOption,
		"keybinding.compare-all":   SwitchCompareLayerListBindingOption,
	}

	hideBindings := map[string]interface{}{
		"keybinding.page-up":   true,
		"keybinding.page-down": true,
	}

	for keybinding, action := range bindingSettings {
		binding, err := config.GetKeyBinding(keybinding)
		if err != nil {
			panic(fmt.Errorf("setup error for keybinding: %s: %w", keybinding, err))
			// TODO handle this error
			//return nil
		}
		_, hidden := hideBindings[keybinding]
		ll.AddBindingOptions(action(KeyBindingDisplay{KeyBinding: &binding, Selected: false, Hide: hidden}))
	}

	return ll
}

func (ll *LayerList) GetKeyBindings() []KeyBindingDisplay {
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
			line = fmt.Sprintf("%s %s", cmpFormatter(cmpString), lineFormatter(fmt.Sprintf("%d", yIndex + 1)))
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

	logrus.Debugln("keyUp in layers")
	logrus.Debugf("  cmpIndex: %d", ll.cmpIndex)
	logrus.Debugf("  bufferIndexLowerBound: %d", ll.bufferIndexLowerBound)
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
	logrus.Debugln("keyDown in layers")
	logrus.Debugf("  cmpIndex: %d", ll.cmpIndex)
	logrus.Debugf("  bufferIndexLowerBound: %d", ll.bufferIndexLowerBound)

	return ll.SetLayerIndex(ll.cmpIndex)
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
