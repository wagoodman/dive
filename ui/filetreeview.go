package ui

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
	"github.com/lunixbochs/vtclean"
	"github.com/wagoodman/docker-image-explorer/filetree"
)

type FileTreeView struct {
	Name            string
	gui             *gocui.Gui
	view            *gocui.View
	header          *gocui.View
	TreeIndex       int
	ModelTree       *filetree.FileTree
	ViewTree        *filetree.FileTree
	RefTrees        []*filetree.FileTree
	HiddenDiffTypes []bool
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
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyCtrlSlash, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return nil }); err != nil {
		return err
	}

	view.updateViewTree()
	view.Render()

	headerStr := fmt.Sprintf(filetree.AttributeFormat+" %s", "P", "ermission", "UID:GID", "Size", "Filetree")
	fmt.Fprintln(view.header, Formatting.Header(vtclean.Clean(headerStr, false)))

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
	var filterBytes []byte
	var filterRegex *regexp.Regexp
	read, err := Views.Command.view.Read(filterBytes)
	if read > 0 && err == nil {
		regex, err := regexp.Compile(string(filterBytes))
		if err == nil {
			filterRegex = regex
		}
	}

	evaluator = func(curNode *filetree.FileNode) bool {
		regexMatch := true
		if filterRegex != nil {
			match := filterRegex.Find([]byte(curNode.Path()))
			regexMatch = match != nil
		}
		return !curNode.Parent.Data.ViewInfo.Collapsed && !curNode.Data.ViewInfo.Hidden && regexMatch
	}

	err = view.ModelTree.VisitDepthParentFirst(visiter, evaluator)
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

func filterRegex() *regexp.Regexp {
	debugPrint("Entered filterRegex()")
	if Views.Command == nil || Views.Command.view == nil {
		return nil
	}
	filterString := strings.TrimSpace(Views.Command.view.Buffer())
	if len(filterString) < 1 {
		debugPrint(fmt.Sprintf("returing nil from fitlerRegex() because string is too short (%s)", filterString))
		return nil
	}

	debugPrint("Compiling regex from " + filterString)
	regex, err := regexp.Compile(filterString)
	if err != nil {
		debugPrint("Returning nil from filterRegex")
		return nil
	}

	return regex
}

func debugPrint(s string) {
	if debug && Views.Tree != nil && Views.Tree.gui != nil {
		v, _ := Views.Tree.gui.View("debug")
		if v != nil {
			if len(v.ViewBuffer()) > 100 {
				v.Clear()
			}
			_, _ = fmt.Fprintln(v, s)
		}
	}
}

func (view *FileTreeView) updateViewTree() {
	regex := filterRegex()
	if regex == nil {
		debugPrint("Nil regex in updateViewTree()")
	}
	// keep the view selection in parity with the current DiffType selection
	view.ModelTree.VisitDepthChildFirst(func(node *filetree.FileNode) error {
		node.Data.ViewInfo.Hidden = view.HiddenDiffTypes[node.Data.DiffType]
		visibleChild := false
		for _, child := range node.Children {
			if !child.Data.ViewInfo.Hidden {
				visibleChild = true
			}
		}
		if regex != nil && !visibleChild {
			match := regex.FindString(node.Path())
			node.Data.ViewInfo.Hidden = len(match) == 0
			debugPrint(fmt.Sprintf("Not nil regex, match was %s", string(match)))
			if len(match) == 0 {
				debugPrint(fmt.Sprintf("Hiding '%s' because of failure to match /%v/", node.Path(), regex))
			} else {
				debugPrint(fmt.Sprintf("Showing '%s' because of matching /%v/", node.Path(), regex))
			}
		}
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
	control := color.New(color.Bold).SprintFunc()
	return control("[Space]") + ": Collapse dir " +
		control("[^A]") + ": Added files " +
		control("[^R]") + ": Removed files " +
		control("[^M]") + ": Modified files " +
		control("[^U]") + ": Unmodified files"
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
