package ui

import (
	"log"

	"github.com/jroimartin/gocui"
	"github.com/wagoodman/docker-image-explorer/filetree"
	"github.com/wagoodman/docker-image-explorer/image"
)

var Views struct {
	Tree  *FileTreeView
	Layer *LayerView
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == Views.Layer.Name {
		_, err := g.SetCurrentView(Views.Tree.Name)
		return err
	}
	_, err := g.SetCurrentView(Views.Layer.Name)
	return err
}

func CursorDown(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()

	// if there isn't a next line
	line, err := v.Line(cy + 1)
	if err != nil {
		// todo: handle error
	}
	if len(line) == 0 {
		return nil
	}
	if err := v.SetCursor(cx, cy+1); err != nil {
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+1); err != nil {
			return err
		}
	}
	return nil
}

func CursorUp(g *gocui.Gui, v *gocui.View) error {
	ox, oy := v.Origin()
	cx, cy := v.Cursor()
	if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
		if err := v.SetOrigin(ox, oy-1); err != nil {
			return err
		}
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	//if err := g.SetKeybinding("main", gocui.MouseLeft, gocui.ModNone, toggleCollapse); err != nil {
	//	return err
	//}
	if err := g.SetKeybinding("side", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}

	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	splitCol := 100
	debugCol := maxX - 70
	if view, err := g.SetView(Views.Layer.Name, -1, -1, splitCol, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		Views.Layer.Setup(view)

	}
	if view, err := g.SetView(Views.Tree.Name, splitCol, -1, debugCol, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		Views.Tree.Setup(view)

		if _, err := g.SetCurrentView(Views.Tree.Name); err != nil {
			return err
		}
	}
	if _, err := g.SetView("debug", debugCol, -1, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}

	return nil
}

func Run(layers []*image.Layer, refTrees []*filetree.FileTree) {

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	Views.Layer = NewLayerView("side", g, layers)
	Views.Tree = NewFileTreeView("main", g, filetree.StackRange(refTrees, 0), refTrees)

	g.Cursor = false
	//g.Mouse = true
	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
