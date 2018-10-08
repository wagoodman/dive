package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/wagoodman/dive/image"
	"github.com/lunixbochs/vtclean"
	"github.com/dustin/go-humanize"
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

func (view *LayerView) IsVisible() bool {
	if view == nil {return false}
	return true
}

func (view *LayerView) setCompareMode(compareMode CompareType) error {
	view.CompareMode = compareMode
	Update()
	Render()
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

func (view *LayerView) Update() error {
	return nil
}

func (view *LayerView) Render() error {
	view.gui.Update(func(g *gocui.Gui) error {
		// update header
		headerStr := fmt.Sprintf("Cmp "+image.LayerFormat, "Image ID", "%Eff.", "Size", "Filter")
		fmt.Fprintln(view.header, Formatting.Header(vtclean.Clean(headerStr, false)))

		// update contents
		view.view.Clear()
		for revIdx := len(view.Layers) - 1; revIdx >= 0; revIdx-- {
			layer := view.Layers[revIdx]
			idx := (len(view.Layers)-1) - revIdx

			layerStr := layer.String()
			if idx == 0 {
				var layerId string
				if len(layer.History.ID) >= 25 {
					layerId = layer.History.ID[0:25]
				} else {
					layerId = fmt.Sprintf("%-25s", layer.History.ID)
				}

				layerStr = fmt.Sprintf(image.LayerFormat, layerId, "", humanize.Bytes(uint64(layer.History.Size)), "FROM "+layer.Id())
			}

			compareBar := view.renderCompareBar(idx)

			if idx == view.LayerIndex {
				fmt.Fprintln(view.view, compareBar + "  " + Formatting.Selected(layerStr))
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
			// debugPrint(fmt.Sprintf("%d",len(filetree.Cache)))
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
			// debugPrint(fmt.Sprintf("%d",len(filetree.Cache)))
		}
	}
	return nil
}

func (view *LayerView) SetCursor(layer int) error {
	// view.view.SetCursor(0, layer)
	view.LayerIndex = layer
	Views.Tree.setTreeByLayer(view.getCompareIndexes())
	view.Render()

	return nil
}

func (view *LayerView) KeyHelp() string {
	return  renderStatusOption("^L","Layer changes", view.CompareMode == CompareLayer) +
			renderStatusOption("^A","All changes", view.CompareMode == CompareAll)
}
