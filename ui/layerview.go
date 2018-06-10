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
	LayerIndex int
	Layers     []*image.Layer
}

func NewLayerView(name string, gui *gocui.Gui, layers []*image.Layer) (layerview *LayerView) {
	layerview = new(LayerView)

	// populate main fields
	layerview.Name = name
	layerview.gui = gui
	layerview.Layers = layers

	return layerview
}

func (view *LayerView) Setup(v *gocui.View) error {

	// set view options
	view.view = v
	view.view.Wrap = false
	view.view.Highlight = true
	view.view.SelBgColor = gocui.ColorGreen
	view.view.SelFgColor = gocui.ColorBlack
	view.view.Frame = false

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
		for idx := len(view.Layers) - 1; idx >= 0; idx-- {
			layer := view.Layers[idx]
			fmt.Fprintln(view.view, layer.String())
		}
		return nil
	})
	// todo: blerg
	return nil
}

func (view *LayerView) CursorDown() error {
	if int(view.LayerIndex) < len(view.Layers) {
		err := CursorDown(view.gui, view.view)
		if err == nil {
			view.LayerIndex++
			view.Render()
			Views.Tree.setLayer(view.LayerIndex)
		}
	}
	return nil
}

func (view *LayerView) CursorUp() error {
	if int(view.LayerIndex) > 0 {
		err := CursorUp(view.gui, view.view)
		if err == nil {
			view.LayerIndex--
			view.Render()
			Views.Tree.setLayer(view.LayerIndex)
		}
	}
	return nil
}
