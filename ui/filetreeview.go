package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/wagoodman/docker-image-explorer/filetree"
	"strings"
	"github.com/lunixbochs/vtclean"
)

const (
	CompareLayer CompareType = iota
	CompareAll
)

type CompareType int


type FileTreeView struct {
	Name              string
	gui               *gocui.Gui
	view              *gocui.View
	header            *gocui.View
	ModelTree         *filetree.FileTree
	ViewTree          *filetree.FileTree
	RefTrees          []*filetree.FileTree
	HiddenDiffTypes   []bool
	TreeIndex         int

}

func NewFileTreeView(name string, gui *gocui.Gui, tree *filetree.FileTree, refTrees []*filetree.FileTree) (treeview *FileTreeView) {
	treeview = new(FileTreeView)

	// populate main fields
	treeview.Name = name
	treeview.gui = gui
	treeview.ModelTree = tree
	treeview.RefTrees = refTrees
	treeview.HiddenDiffTypes = make([]bool, 4)

	return treeview
}

func (view *FileTreeView) Setup(v *gocui.View, header *gocui.View) error {

	// set view options
	view.view = v
	view.view.Editable = false
	view.view.Wrap = false
	//view.view.Highlight = true
	//view.view.SelBgColor = gocui.ColorGreen
	//view.view.SelFgColor = gocui.ColorBlack
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

	view.updateViewTree()
	view.Render()

	headerStr := fmt.Sprintf(filetree.AttributeFormat + " %s", "P","ermission", "UID:GID", "Size", "Filetree")
	fmt.Fprintln(view.header, Formatting.Header(vtclean.Clean(headerStr, false)))

	return nil
}



func (view *FileTreeView) setTreeByLayer(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) error {
	//if stopIdx > len(view.RefTrees)-1 {
	//	return errors.New(fmt.Sprintf("Invalid layer index given: %d of %d", stopIdx, len(view.RefTrees)-1))
	//}
	newTree := filetree.StackRange(view.RefTrees, bottomTreeStart, bottomTreeStop)

	for idx := topTreeStart; idx <= topTreeStop; idx++ {
		newTree.Compare(view.RefTrees[idx])
	}

	// preserve view state on copy
	visitor := func(node *filetree.FileNode) error {
		newNode, err := newTree.GetNode(node.Path())
		if err == nil {
			newNode.Data.ViewInfo = node.Data.ViewInfo
		}
		return nil
	}
	view.ModelTree.VisitDepthChildFirst(visitor, nil)

	view.view.SetCursor(0, 0)
	view.TreeIndex = 0
	view.ModelTree = newTree
	view.updateViewTree()
	return view.Render()
}

func (view *FileTreeView) CursorDown() error {
	// cannot easily (quickly) check the model length, allow the view
	// to let us know what is a valid bounds (i.e. when it hits an empty line)
	err := CursorDown(view.gui, view.view)
	if err == nil {
		view.TreeIndex++
	}
	return view.Render()
}

func (view *FileTreeView) CursorUp() error {
	if view.TreeIndex > 0 {
		err := CursorUp(view.gui, view.view)
		if err == nil {
			view.TreeIndex--
		}
	}
	return view.Render()
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
		return !curNode.Parent.Data.ViewInfo.Collapsed && !curNode.Data.ViewInfo.Hidden
	}

	err := view.ModelTree.VisitDepthParentFirst(visiter, evaluator)
	if err != nil {
		panic(err)
	}

	return node
}

func (view *FileTreeView) toggleCollapse() error {
	node := view.getAbsPositionNode()
	node.Data.ViewInfo.Collapsed = !node.Data.ViewInfo.Collapsed
	view.updateViewTree()
	return view.Render()
}

func (view *FileTreeView) toggleShowDiffType(diffType filetree.DiffType) error {
	view.HiddenDiffTypes[diffType] = !view.HiddenDiffTypes[diffType]

	view.view.SetCursor(0, 0)
	view.TreeIndex = 0
	view.updateViewTree()
	return view.Render()
}

func (view *FileTreeView) updateViewTree() {
	// keep the view selection in parity with the current DiffType selection
	view.ModelTree.VisitDepthChildFirst(func(node *filetree.FileNode) error {
		node.Data.ViewInfo.Hidden = view.HiddenDiffTypes[node.Data.DiffType]
		return nil
	}, nil)

	// make a new tree with only visible nodes
	view.ViewTree = view.ModelTree.Copy()
	view.ViewTree.VisitDepthParentFirst(func(node *filetree.FileNode) error {
		if node.Data.ViewInfo.Hidden {
			view.ViewTree.RemovePath(node.Path())
		}
		return nil
	}, nil)
}

func (view *FileTreeView) KeyHelp() string {
	return  Formatting.Control("[Space]") + ": Collapse dir " +
		Formatting.Control("[^A]") + ": Added files " +
		Formatting.Control("[^R]") + ": Removed files " +
		Formatting.Control("[^M]") + ": Modified files " +
		Formatting.Control("[^U]") + ": Unmodified files"
}

func (view *FileTreeView) Render() error {
	// print the tree to the view
	lines := strings.Split(view.ViewTree.String(true), "\n")
	view.gui.Update(func(g *gocui.Gui) error {
		view.view.Clear()
		for idx, line := range lines {
			if idx == view.TreeIndex {
				fmt.Fprintln(view.view, Formatting.StatusBar(vtclean.Clean(line, false)))
			} else {
				fmt.Fprintln(view.view, line)
			}
		}
		// todo: should we check error on the view println?
		return nil
	})
	return nil
}
