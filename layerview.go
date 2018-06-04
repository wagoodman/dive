package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

type LayerView struct {
	name       string
	gui        *gocui.Gui
	view       *gocui.View
	layerIndex uint
	manifest   *Manifest
}

func NewLayerView(name string, gui *gocui.Gui, view *gocui.View, manifest *Manifest) (layerview *LayerView) {
	layerview = new(LayerView)

	// populate main fields
	layerview.name = name
	layerview.gui = gui
	layerview.view = view
	layerview.manifest = manifest

	// set view options
	layerview.view.Wrap = true
	layerview.view.Highlight = true
	layerview.view.SelBgColor = gocui.ColorGreen
	layerview.view.SelFgColor = gocui.ColorBlack

	layerview.render()

	return layerview
}

func (view *LayerView) keybindings() error {

	if err := view.gui.SetKeybinding("side", gocui.KeyArrowDown, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.cursorDown() }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding("side", gocui.KeyArrowUp, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.cursorUp() }); err != nil {
		return err
	}

	return nil
}

func (view *LayerView) render() error {
	view.gui.Update(func(g *gocui.Gui) error {
		view.view.Clear()
		for ix, layerName := range view.manifest.Layers {
			fmt.Fprintf(view.view, "%d: %s\n", ix+1, layerName[0:25])
		}
		return nil
	})
	// todo: blerg
	return nil
}

func (view *LayerView) cursorDown() error {
	if int(view.layerIndex) < len(data.manifest.Layers) {
		cursorDown(view.gui, view.view)
		view.layerIndex++
		view.render()
		views.treeView.reset(StackRange(data.refTrees, view.layerIndex))
	}
	return nil
}

func (view *LayerView) cursorUp() error {
	if int(view.layerIndex) > 0 {
		cursorUp(view.gui, view.view)
		view.layerIndex--
		view.render()
		views.treeView.reset(StackRange(data.refTrees, view.layerIndex))
	}
	return nil
}
