package ui

import (
	"errors"
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/wagoodman/docker-image-explorer/filetree"
)

type FileTreeView struct {
	Name            string
	gui             *gocui.Gui
	view            *gocui.View
	TreeIndex       int
	Tree            *filetree.FileTree
	RefTrees        []*filetree.FileTree
	HiddenDiffTypes []bool
}

func NewFileTreeView(name string, gui *gocui.Gui, tree *filetree.FileTree, refTrees []*filetree.FileTree) (treeview *FileTreeView) {
	treeview = new(FileTreeView)

	// populate main fields
	treeview.Name = name
	treeview.gui = gui
	treeview.Tree = tree
	treeview.RefTrees = refTrees
	treeview.HiddenDiffTypes = make([]bool, 4)

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
	view.view.Frame = true

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
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyCtrlA, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.toggleShowDiffType(filetree.Added) }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyCtrlR, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.toggleShowDiffType(filetree.Removed) }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyCtrlM, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.toggleShowDiffType(filetree.Changed) }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyCtrlU, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.toggleShowDiffType(filetree.Unchanged) }); err != nil {
		return err
	}

	view.Render()

	return nil
}

func (view *FileTreeView) setLayer(layerIndex int) error {
	if layerIndex > len(view.RefTrees)-1 {
		return errors.New(fmt.Sprintf("Invalid layer index given: %d of %d", layerIndex, len(view.RefTrees)-1))
	}
	newTree := filetree.StackRange(view.RefTrees, layerIndex-1)
	newTree.Compare(view.RefTrees[layerIndex])

	// preserve view state on copy
	visitor := func(node *filetree.FileNode) error {
		newNode, err := newTree.GetNode(node.Path())
		if err == nil {
			newNode.Collapsed = node.Collapsed
		}
		return nil
	}
	view.Tree.Visit(visitor)

	// now that the tree has been rebuilt, keep the view seleciton in parity with the previous selection
	view.setHiddenFromDiffTypes()

	if debug {
		v, _ := view.gui.View("debug")
		v.Clear()
		_, _ = fmt.Fprintln(v, view.RefTrees[layerIndex])
	}

	view.view.SetCursor(0, 0)
	view.TreeIndex = 0
	view.Tree = newTree
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
		return !curNode.Collapsed && !curNode.Hidden
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

func (view *FileTreeView) setHiddenFromDiffTypes() error {
	visitor := func(node *filetree.FileNode) error {
		node.Hidden = view.HiddenDiffTypes[node.Data.DiffType]
		return nil
	}
	view.Tree.Visit(visitor)
	return view.Render()
}

func (view *FileTreeView) toggleShowDiffType(diffType filetree.DiffType) error {
	view.HiddenDiffTypes[diffType] = !view.HiddenDiffTypes[diffType]
	return view.setHiddenFromDiffTypes()
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
