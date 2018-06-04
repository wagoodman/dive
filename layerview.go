package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

func renderSideBar(g *gocui.Gui, v *gocui.View) error {
	g.Update(func(g *gocui.Gui) error {
		v, _ := g.View("side")
		// todo: handle above error.
		v.Clear()
		//_, err := fmt.Fprintf(v, "FileNode:\n%+v\n\n", getAbsPositionNode())
		for ix, layerName := range data.manifest.Layers {
			fmt.Fprintf(v, "%d: %s\n", ix+1, layerName[0:25])
		}
		return nil
	})
	// todo: blerg
	return nil
}

func cursorDownLayers(g *gocui.Gui, v *gocui.View) error {
	if v != nil && int(view.layerIndex) < len(data.manifest.Layers) {
		cursorDown(g, v)
		view.layerIndex++
		renderSideBar(g, v)
		view.treeView.reset(StackRange(data.refTrees, view.layerIndex))
	}
	return nil
}

func cursorUpLayers(g *gocui.Gui, v *gocui.View) error {
	if v != nil && int(view.layerIndex) > 0 {
		cursorUp(g, v)
		view.layerIndex--
		renderSideBar(g, v)
		view.treeView.reset(StackRange(data.refTrees, view.layerIndex))
	}
	return nil
}
