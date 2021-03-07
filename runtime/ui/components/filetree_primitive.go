package components

import (
	"fmt"
	"strings"

	"github.com/wagoodman/dive/runtime/config"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/internal/log"
	"github.com/wagoodman/dive/runtime/ui/components/helpers"
	"github.com/wagoodman/dive/runtime/ui/format"
)

type TreeModel interface {
	StringBetween(int, int, bool) string
	VisitDepthParentFirst(filetree.Visitor, filetree.VisitEvaluator) error
	VisitDepthChildFirst(filetree.Visitor, filetree.VisitEvaluator) error
	RemovePath(path string) error
	VisibleSize() int
	SetLayerIndex(int) bool
	ToggleHiddenFileType(filetype filetree.DiffType) bool
	GetHiddenFileType(filetype filetree.DiffType) bool
}

type TreeViewOption func(t *TreeView)

func upBindingOption() TreeViewOption {
	k := helpers.NewKeyBinding("Cursor Up", tcell.NewEventKey(tcell.KeyUp, rune(0), tcell.ModNone))
	return func(t *TreeView) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   alwaysFalse,
			Hide:       alwaysTrue,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.keyUp() })
	}
}

func downBindingOption() TreeViewOption {
	k := helpers.NewKeyBinding("Cursor Down", tcell.NewEventKey(tcell.KeyDown, rune(0), tcell.ModNone))
	return func(t *TreeView) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   alwaysFalse,
			Hide:       alwaysTrue,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.keyDown() })
	}
}

func rightBindingOption() TreeViewOption {
	k := helpers.NewKeyBinding("Cursor Right", tcell.NewEventKey(tcell.KeyRight, rune(0), tcell.ModNone))
	return func(t *TreeView) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   alwaysFalse,
			Hide:       alwaysTrue,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.keyRight() })
	}
}

func leftBindingOption() TreeViewOption {
	k := helpers.NewKeyBinding("Cursor Left", tcell.NewEventKey(tcell.KeyLeft, rune(0), tcell.ModNone))
	return func(t *TreeView) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   alwaysFalse,
			Hide:       alwaysTrue,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.keyLeft() })
	}
}

func pageUpBindingOption(bindingValue string) TreeViewOption {
	k := helpers.NewKeyBinding("Pg Up", helpers.DecodeBinding(bindingValue))
	return func(t *TreeView) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   alwaysFalse,
			Hide:       alwaysTrue,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.pageUp() })
	}
}

func pageDownBindingOption(bindingValue string) TreeViewOption {
	k := helpers.NewKeyBinding("Pg Down", helpers.DecodeBinding(bindingValue))
	return func(t *TreeView) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   alwaysFalse,
			Hide:       alwaysTrue,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.pageDown() })
	}
}

func collapseDirBindingOption(bindingValue string) TreeViewOption {
	k := helpers.NewKeyBinding("Collapse", helpers.DecodeBinding(bindingValue))
	return func(t *TreeView) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   alwaysFalse,
			Hide:       alwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.collapseDir() })
	}
}

func collapseAllBindingOption(bindingValue string) TreeViewOption {
	k := helpers.NewKeyBinding("Collapse All", helpers.DecodeBinding(bindingValue))
	return func(t *TreeView) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   alwaysFalse,
			Hide:       alwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.CollapseOrExpandAll() })
	}
}

func toggleAttributesOption(bindingValue string) TreeViewOption {
	k := helpers.NewKeyBinding("Attr.", helpers.DecodeBinding(bindingValue))
	return func(t *TreeView) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   func() bool { return t.showAttributes },
			Hide:       alwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.showAttributes = !t.showAttributes })
	}
}

func toggleAddedFilesOption(bindingValue string) TreeViewOption {
	k := helpers.NewKeyBinding("Added", helpers.DecodeBinding(bindingValue))
	return func(t *TreeView) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   func() bool { return !t.tree.GetHiddenFileType(filetree.Added) },
			Hide:       alwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.tree.ToggleHiddenFileType(filetree.Added) })
	}
}

func toggleRemovedFilesOption(bindingValue string) TreeViewOption {
	k := helpers.NewKeyBinding("Removed", helpers.DecodeBinding(bindingValue))
	return func(t *TreeView) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   func() bool { return !t.tree.GetHiddenFileType(filetree.Removed) },
			Hide:       alwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.tree.ToggleHiddenFileType(filetree.Removed) })
	}
}

func toggleModifiedFilesOption(bindingValue string) TreeViewOption {
	k := helpers.NewKeyBinding("Modified", helpers.DecodeBinding(bindingValue))
	return func(t *TreeView) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   func() bool { return !t.tree.GetHiddenFileType(filetree.Modified) },
			Hide:       alwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.tree.ToggleHiddenFileType(filetree.Modified) })

	}
}

func toggleUnmodifiedFilesOption(bindingValue string) TreeViewOption {
	k := helpers.NewKeyBinding("Unmodified", helpers.DecodeBinding(bindingValue))
	return func(t *TreeView) {
		displayBinding := helpers.KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   func() bool { return !t.tree.GetHiddenFileType(filetree.Unmodified) },
			Hide:       alwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.tree.ToggleHiddenFileType(filetree.Unmodified) })
	}
}

type TreeView struct {
	*tview.Box
	tree TreeModel
	// Note that the following two fields are distinct
	// treeIndex is the index about where we are in the current fileTree
	// this should be updated every keypress
	treeIndex             int
	bufferIndexLowerBound int
	globalCollapseAll     bool
	keyInputHandler       *helpers.KeyInputHandler
	showAttributes        bool
}

func NewTreeView(tree TreeModel) *TreeView {
	return &TreeView{
		Box:               tview.NewBox(),
		tree:              tree,
		globalCollapseAll: true,
		showAttributes:    true,
		keyInputHandler:   helpers.NewKeyInputHandler(),
	}
}

func (t *TreeView) AddBindingOptions(bindingOptions ...TreeViewOption) *TreeView {
	for _, option := range bindingOptions {
		option(t)
	}

	return t
}

// Implementation notes:
//need to set up our input handler here,
//Should probably factor out keybinding initialization into a new function

func (t *TreeView) Setup(cfg config.KeybindingConfig) *TreeView {
	t.tree.SetLayerIndex(0)

	t.AddBindingOptions(
		upBindingOption(),
		downBindingOption(),
		leftBindingOption(),
		rightBindingOption(),
		collapseDirBindingOption(cfg.ToggleCollapseDir),
		collapseAllBindingOption(cfg.ToggleCollapseAllDir),
		toggleAddedFilesOption(cfg.ToggleAddedFiles),
		toggleRemovedFilesOption(cfg.ToggleRemovedFiles),
		toggleModifiedFilesOption(cfg.ToggleModifiedFiles),
		toggleUnmodifiedFilesOption(cfg.ToggleUnmodifiedFiles),
		toggleAttributesOption(cfg.ToggleFileTreeAttributes),
		pageUpBindingOption(cfg.PageUp),
		pageDownBindingOption(cfg.PageDown),
	)
	return t
}

// TODO: do we need all of these?? or is there an alternative API we could use for the wrappers????
func (t *TreeView) getBox() *tview.Box {
	return t.Box
}

func (t *TreeView) getDraw() drawFn {
	return t.Draw
}

func (t *TreeView) getInputWrapper() inputFn {
	return t.InputHandler
}

// Keybinding list

func (t *TreeView) GetKeyBindings() []helpers.KeyBindingDisplay {
	return t.keyInputHandler.Order
}

func (t *TreeView) ToggleHideAttributes() {
	t.showAttributes = !t.showAttributes
}

// Implementation note:
// what do we want here??? a binding object?? yes
func (t *TreeView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return t.keyInputHandler.Handle()
}

func (t *TreeView) WrapInputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return t.Box.WrapInputHandler(t.InputHandler())
}

func (t *TreeView) Focus(delegate func(p tview.Primitive)) {
	t.Box.Focus(delegate)
}

func (t *TreeView) HasFocus() bool {
	return t.Box.HasFocus()
}

// Private helper methods

func (t *TreeView) collapseDir() bool {
	node := t.getAbsPositionNode()
	if node != nil && node.Data.FileInfo.IsDir {
		log.WithFields(
			"path", node.Path(),
		).Trace("collapsing node")
		node.Data.ViewInfo.Collapsed = !node.Data.ViewInfo.Collapsed
	}
	return true
}

func (t *TreeView) CollapseOrExpandAll() bool {
	visitor := func(n *filetree.FileNode) error {
		if n != nil && n.Data.FileInfo.IsDir {
			n.Data.ViewInfo.Collapsed = t.globalCollapseAll
		}
		return nil
	}

	evaluator := func(n *filetree.FileNode) bool {
		return true
	}
	if err := t.tree.VisitDepthParentFirst(visitor, evaluator); err != nil {
		err = fmt.Errorf("error collapsing all directories: %w", err)
		log.Error(err)
		panic(err)
	}

	t.globalCollapseAll = !t.globalCollapseAll

	return true

}

// getAbsPositionNode determines the selected screen cursor's location in the file tree, returning the selected FileNode.
func (t *TreeView) getAbsPositionNode() (node *filetree.FileNode) {
	var visitor func(*filetree.FileNode) error
	var evaluator func(*filetree.FileNode) bool
	var dfsCounter int

	visitor = func(curNode *filetree.FileNode) error {
		if dfsCounter == t.treeIndex {
			node = curNode
		}
		dfsCounter++
		return nil
	}

	evaluator = func(curNode *filetree.FileNode) bool {
		return !curNode.Parent.Data.ViewInfo.Collapsed && !curNode.Data.ViewInfo.Hidden
	}

	err := t.tree.VisitDepthParentFirst(visitor, evaluator)
	if err != nil {
		log.Errorf("unable to get node position: %+v", err)
	}

	return node
}

func (t *TreeView) GetInnerRect() (int, int, int, int) {
	x, y, width, height := t.Box.GetInnerRect()
	return x, y + 1, width, height - 1
}

func (t *TreeView) keyDown() bool {
	_, _, _, height := t.GetInnerRect()

	// treeIndex is the index about where we are in the current file
	if t.treeIndex >= t.tree.VisibleSize() {
		return false
	}
	t.treeIndex++
	if (t.treeIndex - t.bufferIndexLowerBound) >= height {
		t.bufferIndexLowerBound++
	}

	log.WithFields(
		"component", "TreeView",
		"path", t.getAbsPositionNode().Path(),
		"treeIndex", t.treeIndex,
		"bufferIndexLowerBound", t.bufferIndexLowerBound,
		"height", height,
	).Tracef("keyDown event")
	return true
}

func (t *TreeView) keyUp() bool {
	if t.treeIndex <= 0 {
		return false
	}
	t.treeIndex--
	if t.treeIndex < t.bufferIndexLowerBound {
		t.bufferIndexLowerBound--
	}

	log.WithFields(
		"component", "TreeView",
		"path", t.getAbsPositionNode().Path(),
		"treeIndex", t.treeIndex,
		"bufferIndexLowerBound", t.bufferIndexLowerBound,
	).Tracef("keyUp event")
	return true
}

// TODO add regex filtering
func (t *TreeView) keyRight() bool {
	node := t.getAbsPositionNode()

	_, _, _, height := t.GetInnerRect()
	if node == nil {
		return false
	}

	if !node.Data.FileInfo.IsDir {
		return false
	}

	if len(node.Children) == 0 {
		return false
	}

	if node.Data.ViewInfo.Collapsed {
		node.Data.ViewInfo.Collapsed = false
	}

	t.treeIndex++
	if (t.treeIndex - t.bufferIndexLowerBound) >= height {
		t.bufferIndexLowerBound++
	}

	return true
}

func (t *TreeView) keyLeft() bool {
	var visitor func(*filetree.FileNode) error
	var evaluator func(*filetree.FileNode) bool
	var dfsCounter, newIndex int
	//oldIndex := t.treeIndex
	currentNode := t.getAbsPositionNode()

	if currentNode == nil {
		return true
	}
	parentPath := currentNode.Parent.Path()

	visitor = func(curNode *filetree.FileNode) error {
		if strings.Compare(parentPath, curNode.Path()) == 0 {
			newIndex = dfsCounter
		}
		dfsCounter++
		return nil
	}

	evaluator = func(curNode *filetree.FileNode) bool {
		return !curNode.Parent.Data.ViewInfo.Collapsed && !curNode.Data.ViewInfo.Hidden
	}

	err := t.tree.VisitDepthParentFirst(visitor, evaluator)
	if err != nil {
		// TODO: remove this panic
		panic(err)
	}

	t.treeIndex = newIndex
	//moveIndex := oldIndex - newIndex
	if newIndex < t.bufferIndexLowerBound {
		t.bufferIndexLowerBound = t.treeIndex
	}

	return true
}

// TODO make all movement rely on a single function (shouldn't be too dificult really)
func (t *TreeView) pageDown() bool {

	_, _, _, height := t.GetInnerRect()
	visibleSize := t.tree.VisibleSize()
	t.treeIndex = intMin(t.treeIndex+height, visibleSize)
	if t.treeIndex >= t.bufferIndexUpperBound() {
		t.bufferIndexLowerBound = intMin(t.treeIndex, visibleSize-height+1)
	}
	return true
}

func (t *TreeView) pageUp() bool {
	_, _, _, height := t.GetInnerRect()

	t.treeIndex = intMax(0, t.treeIndex-height)
	if t.treeIndex < t.bufferIndexLowerBound {
		t.bufferIndexLowerBound = t.treeIndex
	}

	return true
}

func (t *TreeView) bufferIndexUpperBound() int {
	_, _, _, height := t.GetInnerRect()
	return t.bufferIndexLowerBound + height
}

func (t *TreeView) Draw(screen tcell.Screen) {
	t.Box.Draw(screen)
	selectedIndex := t.treeIndex - t.bufferIndexLowerBound
	x, y, width, _ := t.Box.GetInnerRect()
	showAttributes := width > 50 && t.showAttributes
	treeString := t.tree.StringBetween(t.bufferIndexLowerBound, t.bufferIndexUpperBound(), showAttributes)
	lines := strings.Split(treeString, "\n")

	headerLine := "Filetree"
	if showAttributes {
		headerLine = fmt.Sprintf("Permission %11s %10s %s", "UID:GID", "Size", "Filetree")
	}

	format.PrintLine(screen, headerLine, x, y, len(headerLine), tview.AlignLeft, tcell.StyleDefault)
	x, y, _, height := t.GetInnerRect()
	// update the contents
	for yIndex, line := range lines {
		if yIndex >= height {
			break
		}
		lineStyle := tcell.StyleDefault
		lineFormatter := format.None
		if yIndex == selectedIndex {
			lineFormatter = format.BoldReplace
			lineStyle = format.SelectedStyle
		}
		tview.PrintWithStyle(screen, lineFormatter(line), x, y+yIndex, len(line), tview.AlignLeft, lineStyle)

	}

}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
