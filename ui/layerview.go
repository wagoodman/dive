package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/wagoodman/docker-image-explorer/image"
)

type LayerView struct {
	Name       string
	gui        *gocui.Gui
	view       *gocui.View
	LayerIndex uint
	Manifest   *image.Manifest
}

func NewLayerView(name string, gui *gocui.Gui, manifest *image.Manifest) (layerview *LayerView) {
	layerview = new(LayerView)

	// populate main fields
	layerview.Name = name
	layerview.gui = gui
	layerview.Manifest = manifest

	return layerview
}

func (view *LayerView) Setup(v *gocui.View) error {

	// set view options
	view.view = v
	view.view.Wrap = true
	view.view.Highlight = true
	view.view.SelBgColor = gocui.ColorGreen
	view.view.SelFgColor = gocui.ColorBlack

	// set keybindings
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowDown, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorDown() }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowUp, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorUp() }); err != nil {
		return err
	}

	view.Render()

	return nil
}

func (view *LayerView) Render() error {
	view.gui.Update(func(g *gocui.Gui) error {
		view.view.Clear()
		for ix, layerName := range view.Manifest.Layers {
			fmt.Fprintf(view.view, "%d: %s\n", ix+1, layerName[0:25])
		}
		return nil
	})
	// todo: blerg
	return nil
}

func (view *LayerView) CursorDown() error {
	if int(view.LayerIndex) < len(view.Manifest.Layers) {
		CursorDown(view.gui, view.view)
		view.LayerIndex++
		view.Render()
		Views.Tree.setLayer(view.LayerIndex)
	}
	return nil
}

func (view *LayerView) CursorUp() error {
	if int(view.LayerIndex) > 0 {
		CursorUp(view.gui, view.view)
		view.LayerIndex--
		view.Render()
		// this line is evil
		Views.Tree.setLayer(view.LayerIndex)
	}
	return nil
}
