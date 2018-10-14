package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/wagoodman/dive/filetree"
	"github.com/lunixbochs/vtclean"
)

const (
	CompareLayer CompareType = iota
	CompareAll
)

type CompareType int


type FileTreeView struct {
	Name                  string
	gui                   *gocui.Gui
	view                  *gocui.View
	header                *gocui.View
	ModelTree             *filetree.FileTree
	ViewTree              *filetree.FileTree
	RefTrees              []*filetree.FileTree
	HiddenDiffTypes       []bool
	TreeIndex             uint
	bufferIndex           uint
	bufferIndexUpperBound uint
	bufferIndexLowerBound uint
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

	view.bufferIndexLowerBound = 0
	view.bufferIndexUpperBound = view.height() // don't include the header or footer in the view size

	view.Update()
	view.Render()

	return nil
}

func (view *FileTreeView) height() uint {
	_, height := view.view.Size()
	return uint(height - 2)
}

func (view *FileTreeView) IsVisible() bool {
	if view == nil {return false}
	return true
}

func (view *FileTreeView) resetCursor() {
	view.view.SetCursor(0, 0)
	view.TreeIndex = 0
	view.bufferIndex = 0
	view.bufferIndexLowerBound = 0
	view.bufferIndexUpperBound = view.height()
}

func (view *FileTreeView) setTreeByLayer(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) error {
	if topTreeStop > len(view.RefTrees)-1 {
		return fmt.Errorf("Invalid layer index given: %d of %d", topTreeStop, len(view.RefTrees)-1)
	}
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

	view.resetCursor()

	view.ModelTree = newTree
	view.Update()
	return view.Render()
}

func (view *FileTreeView) doCursorUp() {
	view.TreeIndex--
	if view.TreeIndex < view.bufferIndexLowerBound {
		view.bufferIndexUpperBound--
		view.bufferIndexLowerBound--
	}

	if view.bufferIndex > 0 {
		view.bufferIndex--
	}
}

func (view *FileTreeView) doCursorDown() {
	view.TreeIndex++
	if view.TreeIndex > view.bufferIndexUpperBound {
		view.bufferIndexUpperBound++
		view.bufferIndexLowerBound++
	}
	view.bufferIndex++
	if view.bufferIndex > view.height() {
		view.bufferIndex = view.height()
	}
}

func (view *FileTreeView) CursorDown() error {
	// we cannot use the gocui buffer since any state change requires writing the entire tree to the buffer.
	// Instead we are keeping an upper and lower bounds of the tree string to render and only flushing
	// this range into the view buffer. This is much faster when tree sizes are large.

	view.doCursorDown()
	return view.Render()
}



func (view *FileTreeView) CursorUp() error {
	// we cannot use the gocui buffer since any state change requires writing the entire tree to the buffer.
	// Instead we are keeping an upper and lower bounds of the tree string to render and only flushing
	// this range into the view buffer. This is much faster when tree sizes are large.

	if view.TreeIndex > 0 {
		view.doCursorUp()
		return view.Render()
	}
	return nil
}

func (view *FileTreeView) getAbsPositionNode() (node *filetree.FileNode) {
	var visiter func(*filetree.FileNode) error
	var evaluator func(*filetree.FileNode) bool
	var dfsCounter uint

	visiter = func(curNode *filetree.FileNode) error {
		if dfsCounter == view.TreeIndex {
			node = curNode
		}
		dfsCounter++
		return nil
	}
	var filterBytes []byte
	var filterRegex *regexp.Regexp
	read, err := Views.Filter.view.Read(filterBytes)
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
	if node != nil {
		node.Data.ViewInfo.Collapsed = !node.Data.ViewInfo.Collapsed
	}
	view.Update()
	return view.Render()
}

func (view *FileTreeView) toggleShowDiffType(diffType filetree.DiffType) error {
	view.HiddenDiffTypes[diffType] = !view.HiddenDiffTypes[diffType]

	view.resetCursor()

	Update()
	Render()
	return nil
}

func filterRegex() *regexp.Regexp {
	if Views.Filter == nil || Views.Filter.view == nil {
		return nil
	}
	filterString := strings.TrimSpace(Views.Filter.view.Buffer())
	if len(filterString) < 1 {
		return nil
	}

	regex, err := regexp.Compile(filterString)
	if err != nil {
		return nil
	}

	return regex
}

func (view *FileTreeView) Update() error {
	regex := filterRegex()

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
	return nil
}

func (view *FileTreeView) KeyHelp() string {
	return  renderStatusOption("Space","Collapse dir", false) +
			renderStatusOption("^A","Added files", !view.HiddenDiffTypes[filetree.Added]) +
			renderStatusOption("^R","Removed files", !view.HiddenDiffTypes[filetree.Removed]) +
			renderStatusOption("^M","Modified files", !view.HiddenDiffTypes[filetree.Changed]) +
			renderStatusOption("^U","Unmodified files", !view.HiddenDiffTypes[filetree.Unchanged])
}

func (view *FileTreeView) Render() error {
	treeString := view.ViewTree.StringBetween(view.bufferIndexLowerBound, view.bufferIndexUpperBound,true)
	lines := strings.Split(treeString, "\n")

	// undo a cursor down that has gone past bottom of the visible tree
	if view.bufferIndex >= uint(len(lines))-1 {
		view.doCursorUp()
	}

	title := "Current Layer Contents"
	if Views.Layer.CompareMode == CompareAll {
		title = "Aggregated Layer Contents"
	}

	// indicate when selected
	if view.gui.CurrentView() == view.view {
		title = "● "+title
	}

	view.gui.Update(func(g *gocui.Gui) error {
		// update the header
		view.header.Clear()
		width, _ := g.Size()
		headerStr := fmt.Sprintf("[%s]%s\n", title, strings.Repeat("─", width*2))
		headerStr += fmt.Sprintf(filetree.AttributeFormat+" %s", "P", "ermission", "UID:GID", "Size", "Filetree")
		fmt.Fprintln(view.header, Formatting.Header(vtclean.Clean(headerStr, false)))

		// update the contents
		view.view.Clear()
		for idx, line := range lines {
			if uint(idx) == view.bufferIndex {
				fmt.Fprintln(view.view, Formatting.Selected(vtclean.Clean(line, false)))
			} else {
				fmt.Fprintln(view.view, line)
			}
		}
		// todo: should we check error on the view println?
		return nil
	})
	return nil
}
