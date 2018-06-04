package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

type FileTreeView struct {
	name         string
	gui          *gocui.Gui
	view         *gocui.View
	absTreeIndex uint
	tree         *FileTree
}

func NewFileTreeView(name string, gui *gocui.Gui, view *gocui.View, tree *FileTree) (treeview *FileTreeView) {
	treeview = new(FileTreeView)

	// populate main fields
	treeview.name = name
	treeview.gui = gui
	treeview.view = view
	treeview.tree = tree

	// set view options
	treeview.view.Editable = false
	treeview.view.Wrap = false
	treeview.view.Highlight = true
	treeview.view.SelBgColor = gocui.ColorGreen
	treeview.view.SelFgColor = gocui.ColorBlack

	treeview.render()

	return treeview
}

func (view *FileTreeView) keybindings() error {
	if err := view.gui.SetKeybinding(view.name, gocui.KeyArrowDown, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.cursorDown() }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.name, gocui.KeyArrowUp, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.cursorUp() }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.name, gocui.KeySpace, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.toggleCollapse() }); err != nil {
		return err
	}
	return nil
}

// Mehh, this is just a bad method
func (view *FileTreeView) reset(tree *FileTree) error {
	view.tree = tree
	view.view.SetCursor(0, 0)
	view.absTreeIndex = 0
	return view.render()
}

func (view *FileTreeView) cursorDown() error {
	err := cursorDown(view.gui, view.view)
	if err == nil {
		view.absTreeIndex++
	}
	return nil
}

func (view *FileTreeView) cursorUp() error {
	err := cursorUp(view.gui, view.view)
	if err == nil {
		view.absTreeIndex--
	}
	return nil
}

func (view *FileTreeView) getAbsPositionNode() (node *FileNode) {
	var visiter func(*FileNode) error
	var evaluator func(*FileNode) bool
	var dfsCounter uint

	visiter = func(curNode *FileNode) error {
		if dfsCounter == view.absTreeIndex {
			node = curNode
		}
		dfsCounter++
		return nil
	}

	evaluator = func(curNode *FileNode) bool {
		return !curNode.collapsed
	}

	err := view.tree.VisitDepthParentFirst(visiter, evaluator)
	if err != nil {
		// todo: you guessed it, check errors
	}

	return node
}

func (view *FileTreeView) toggleCollapse() error {
	node := view.getAbsPositionNode()
	node.collapsed = !node.collapsed
	return view.render()
}

func (view *FileTreeView) render() error {
	renderString := view.tree.String()
	view.gui.Update(func(g *gocui.Gui) error {
		view.view.Clear()
		_, err := fmt.Fprintln(view.view, renderString)
		return err
	})
	return nil
}
