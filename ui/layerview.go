package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/wagoodman/docker-image-explorer/image"
	"github.com/lunixbochs/vtclean"
)

type LayerView struct {
	Name       string
	gui        *gocui.Gui
	view       *gocui.View
	header     *gocui.View
	LayerIndex int
	Layers     []*image.Layer
	CompareMode       CompareType
	CompareStartIndex int
}

func NewLayerView(name string, gui *gocui.Gui, layers []*image.Layer) (layerview *LayerView) {
	layerview = new(LayerView)

	// populate main fields
	layerview.Name = name
	layerview.gui = gui
	layerview.Layers = layers
	layerview.CompareMode = CompareLayer

	return layerview
}

func (view *LayerView) Setup(v *gocui.View, header *gocui.View) error {

	// set view options
	view.view = v
	view.view.Editable = false
	view.view.Wrap = false
	//view.view.Highlight = true
	//view.view.SelBgColor = gocui.ColorGreen
	//view.view.SelFgColor = gocui.ColorBlack
	view.view.Frame = false

	view.header = header
	view.header.Editable = false
	view.header.Wrap = false
	view.header.Frame = false

	// set keybindings
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowDown, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorDown() }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowUp, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorUp() }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyCtrlL, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.setCompareMode(CompareLayer) }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyCtrlA, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.setCompareMode(CompareAll) }); err != nil {
		return err
	}

	return view.Render()
}

func (view *LayerView) setCompareMode(compareMode CompareType) error {
	view.CompareMode = compareMode
	view.Render()
	return Views.Tree.setTreeByLayer(view.getCompareIndexes())
}

func (view *LayerView) getCompareIndexes() (bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) {
	bottomTreeStart = view.CompareStartIndex
	topTreeStop = view.LayerIndex

	if view.LayerIndex == view.CompareStartIndex {
		bottomTreeStop = view.LayerIndex
		topTreeStart = view.LayerIndex
	} else if view.CompareMode == CompareLayer {
		bottomTreeStop = view.LayerIndex -1
		topTreeStart = view.LayerIndex
	} else {
		bottomTreeStop = view.CompareStartIndex
		topTreeStart = view.CompareStartIndex+1
	}

	return bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop
}

func (view *LayerView) renderCompareBar(layerIdx int) string {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop := view.getCompareIndexes()
	result := "  "

	//if debug {
	//	v, _ := view.gui.View("debug")
	//	v.Clear()
	//	_, _ = fmt.Fprintf(v, "bStart: %d bStop: %d tStart: %d tStop: %d", bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop)
	//}

	if layerIdx >= bottomTreeStart && layerIdx <= bottomTreeStop {
		result = Formatting.CompareBottom("  ")
	}
	if layerIdx >= topTreeStart && layerIdx <= topTreeStop {
		result = Formatting.CompareTop("  ")
	}

	//if bottomTreeStop == topTreeStart {
	//	result += "  "
	//} else {
	//	if layerIdx == bottomTreeStop {
	//		result += "─┐"
	//	} else if layerIdx == topTreeStart {
	//		result += "─┘"
	//	} else {
	//		result += "  "
	//	}
	//}

	return result
}

func (view *LayerView) Render() error {
	view.gui.Update(func(g *gocui.Gui) error {
		// update header
		headerStr := fmt.Sprintf("Cmp "+image.LayerFormat, "Image ID", "Size", "Command")
		fmt.Fprintln(view.header, Formatting.Header(vtclean.Clean(headerStr, false)))

		// update contents
		view.view.Clear()
		for revIdx := len(view.Layers) - 1; revIdx >= 0; revIdx-- {
			layer := view.Layers[revIdx]
			idx := (len(view.Layers)-1) - revIdx

			layerStr := layer.String()
			if idx == 0 {
				// TODO: add size
				layerStr = fmt.Sprintf(image.LayerFormat, layer.History.ID[0:25], "", "FROM "+layer.Id())
			}

			compareBar := view.renderCompareBar(idx)

			if idx == view.LayerIndex {
				fmt.Fprintln(view.view, compareBar + "  " + Formatting.StatusBar(layerStr))
			} else {
				fmt.Fprintln(view.view, compareBar + "  " + layerStr)
			}

		}
		return nil
	})
	// todo: blerg
	return nil
}

func (view *LayerView) CursorDown() error {
	if view.LayerIndex < len(view.Layers) {
		err := CursorDown(view.gui, view.view)
		if err == nil {
			view.LayerIndex++
			Views.Tree.setTreeByLayer(view.getCompareIndexes())
			view.Render()
		}
	}
	return nil
}

func (view *LayerView) CursorUp() error {
	if view.LayerIndex > 0 {
		err := CursorUp(view.gui, view.view)
		if err == nil {
			view.LayerIndex--
			Views.Tree.setTreeByLayer(view.getCompareIndexes())
			view.Render()
		}
	}
	return nil
}

func (view *LayerView) KeyHelp() string {
	return  Formatting.Control("[^L]") + ": Layer Changes " +
		Formatting.Control("[^A]") + ": All Changes "
}
