package components

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

type LayersViewModel interface {
	SetLayerIndex(int) bool
	GetPrintableLayers() []fmt.Stringer
}

type LayerList struct {
	*tview.Box
	bufferIndexLowerBound int
	cmpIndex              int
	changed               LayerListHandler
	LayersViewModel
}

type LayerListHandler func(index int, shortcut rune)

func NewLayerList(model LayersViewModel) *LayerList {
	return &LayerList{
		Box:             tview.NewBox(),
		cmpIndex:        0,
		LayersViewModel: model,
	}
}

func (ll *LayerList) Setup() *LayerList {
	return ll
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
		var cmpColor tcell.Color
		switch {
		case yIndex == ll.cmpIndex:
			cmpColor = tcell.ColorRed
		case yIndex < ll.cmpIndex:
			cmpColor = tcell.ColorBlue
		default:
			cmpColor = tcell.ColorDefault
		}
		line := fmt.Sprintf("%s[white] %s", cmpString, layer)
		tview.Print(screen, line, x, y+yIndex, width, tview.AlignLeft, cmpColor)
		for xIndex := 0; xIndex < width; xIndex++ {
			m, c, style, _ := screen.GetContent(x+xIndex, y+yIndex)
			fg, bg, _ := style.Decompose()
			style = style.Background(fg).Foreground(bg)
			switch {
			case yIndex == ll.cmpIndex:
				screen.SetContent(x+xIndex, y+yIndex, m, c, style)
				screen.SetContent(x+xIndex, y+yIndex, m, c, style)
			case yIndex < ll.cmpIndex && xIndex < len(cmpString):
				screen.SetContent(x+xIndex, y+yIndex, m, c, style)
				screen.SetContent(x+xIndex, y+yIndex, m, c, style)
			default:
				break
			}
		}

	}
}

func (ll *LayerList) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return ll.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp:
			if ll.SetLayerIndex(ll.cmpIndex - 1) {
				ll.keyUp()
				//ll.cmpIndex--
				//logrus.Debugf("KeyUp pressed, index: %d", ll.cmpIndex)
			}
		case tcell.KeyDown:
			if ll.SetLayerIndex(ll.cmpIndex + 1) {
				ll.keyDown()
				//ll.cmpIndex++
				//logrus.Debugf("KeyUp pressed, index: %d", ll.cmpIndex)

			}
		}
	})
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

func (ll *LayerList) keyDown() bool {
	_, _, _, height := ll.Box.GetInnerRect()
	adjustedHeight := height - 1

	// treeIndex is the index about where we are in the current file
	visibleSize := len(ll.GetPrintableLayers())
	if ll.cmpIndex+1+ll.bufferIndexLowerBound >= visibleSize {
		return false
	}
	if ll.cmpIndex+1 >= adjustedHeight {
		ll.bufferIndexLowerBound++
	} else {
		ll.cmpIndex++
	}
	logrus.Debugln("keyDown in layers")
	logrus.Debugf("  cmpIndex: %d", ll.cmpIndex)
	logrus.Debugf("  bufferIndexLowerBound: %d", ll.bufferIndexLowerBound)

	return true
}
