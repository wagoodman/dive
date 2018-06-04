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

func (treeview *FileTreeView) keybindings() error {
	if err := treeview.gui.SetKeybinding(treeview.name, gocui.KeyArrowDown, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return treeview.cursorDown() }); err != nil {
		return err
	}
	if err := treeview.gui.SetKeybinding(treeview.name, gocui.KeyArrowUp, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return treeview.cursorUp() }); err != nil {
		return err
	}
	if err := treeview.gui.SetKeybinding(treeview.name, gocui.KeySpace, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return treeview.toggleCollapse() }); err != nil {
		return err
	}
	return nil
}

// Mehh, this is just a bad method
func (treeview *FileTreeView) reset(tree *FileTree) error {
	treeview.tree = tree
	treeview.view.SetCursor(0, 0)
	treeview.absTreeIndex = 0
	return treeview.render()
}

func (treeview *FileTreeView) cursorDown() error {
	err := cursorDown(treeview.gui, treeview.view)
	if err == nil {
		treeview.absTreeIndex++
	}
	return nil
}

func (treeview *FileTreeView) cursorUp() error {
	err := cursorUp(treeview.gui, treeview.view)
	if err == nil {
		treeview.absTreeIndex--
	}
	return nil
}

func (treeview *FileTreeView) getAbsPositionNode() (node *FileNode) {
	var visiter func(*FileNode) error
	var evaluator func(*FileNode) bool
	var dfsCounter uint

	visiter = func(curNode *FileNode) error {
		if dfsCounter == treeview.absTreeIndex {
			node = curNode
		}
		dfsCounter++
		return nil
	}

	evaluator = func(curNode *FileNode) bool {
		return !curNode.collapsed
	}

	err := treeview.tree.VisitDepthParentFirst(visiter, evaluator)
	if err != nil {
		// todo: you guessed it, check errors
	}

	return node
}

func (treeview *FileTreeView) toggleCollapse() error {
	node := treeview.getAbsPositionNode()
	node.collapsed = !node.collapsed
	return treeview.render()
}

func (treeview *FileTreeView) render() error {
	renderString := treeview.tree.String()
	treeview.gui.Update(func(g *gocui.Gui) error {
		treeview.view.Clear()
		_, err := fmt.Fprintln(treeview.view, renderString)
		return err
	})
	return nil
}
