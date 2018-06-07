package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/wagoodman/docker-image-explorer/filetree"
)

type FileTreeView struct {
	Name      string
	gui       *gocui.Gui
	view      *gocui.View
	TreeIndex int
	Tree      *filetree.FileTree
	RefTrees  []*filetree.FileTree
}

func NewFileTreeView(name string, gui *gocui.Gui, tree *filetree.FileTree, refTrees []*filetree.FileTree) (treeview *FileTreeView) {
	treeview = new(FileTreeView)

	// populate main fields
	treeview.Name = name
	treeview.gui = gui
	treeview.Tree = tree
	treeview.RefTrees = refTrees

	return treeview
}

func (view *FileTreeView) Setup(v *gocui.View) error {

	// set view options
	view.view = v
	view.view.Editable = false
	view.view.Wrap = false
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
	if err := view.gui.SetKeybinding(view.Name, gocui.KeySpace, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.toggleCollapse() }); err != nil {
		return err
	}

	view.Render()

	return nil
}

func (view *FileTreeView) setLayer(layerIndex int) error {
	view.Tree = filetree.StackRange(view.RefTrees, layerIndex-1)
	view.Tree.Compare(view.RefTrees[layerIndex])
	// v, _ := view.gui.View("debug")
	// v.Clear()
	// _, _ = fmt.Fprintln(v, view.RefTrees[layerIndex])
	view.view.SetCursor(0, 0)
	view.TreeIndex = 0
	return view.Render()
}

func (view *FileTreeView) CursorDown() error {
	err := CursorDown(view.gui, view.view)
	if err == nil {
		view.TreeIndex++
	}
	return nil
}

func (view *FileTreeView) CursorUp() error {
	err := CursorUp(view.gui, view.view)
	if err == nil {
		view.TreeIndex--
	}
	return nil
}

func (view *FileTreeView) getAbsPositionNode() (node *filetree.FileNode) {
	var visiter func(*filetree.FileNode) error
	var evaluator func(*filetree.FileNode) bool
	var dfsCounter int

	visiter = func(curNode *filetree.FileNode) error {
		if dfsCounter == view.TreeIndex {
			node = curNode
		}
		dfsCounter++
		return nil
	}

	evaluator = func(curNode *filetree.FileNode) bool {
		return !curNode.Collapsed
	}

	err := view.Tree.VisitDepthParentFirst(visiter, evaluator)
	if err != nil {
		// todo: you guessed it, check errors
	}

	return node
}

func (view *FileTreeView) toggleCollapse() error {
	node := view.getAbsPositionNode()
	node.Collapsed = !node.Collapsed
	return view.Render()
}

func (view *FileTreeView) Render() error {
	renderString := view.Tree.String()
	view.gui.Update(func(g *gocui.Gui) error {
		view.view.Clear()
		_, err := fmt.Fprintln(view.view, renderString)
		return err
	})
	return nil
}
