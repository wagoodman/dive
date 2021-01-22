package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
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

type inputHandleFunc func(event *tcell.EventKey, setFocus func(p tview.Primitive))

// TODO factor out KeyInputHandler and related structs into a separate file
type KeyInputHandler struct {
	Order []KeyBindingDisplay
	HandlerMap map[*tcell.EventKey] func()
}

func NewKeyInputHandler() *KeyInputHandler {
	return &KeyInputHandler{
		Order: []KeyBindingDisplay{},
		HandlerMap: map[*tcell.EventKey] func(){},
	}
}

func (k *KeyInputHandler) AddBinding(binding KeyBindingDisplay, f func() ) *KeyInputHandler {
	k.Order = append(k.Order, binding)
	k.HandlerMap[binding.EventKey] = f

	return k
}

func (k *KeyInputHandler) Handle() inputHandleFunc {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive) ) {
		for _, m := range k.Order {
			if m.Match(event) {
				k.HandlerMap[m.EventKey]()
			}
		}
	}
}

type TreeViewOption func(t *TreeView)

func UpBindingOption(k KeyBinding) TreeViewOption {
	return func (t *TreeView) {
		displayBinding := KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   AlwaysFalse,
			Hide:       AlwaysTrue,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() {t.keyUp()} )
	}
}

func DownBindingOption(k KeyBinding) TreeViewOption {
	return func (t *TreeView) {
		displayBinding := KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   AlwaysFalse,
			Hide:       AlwaysTrue,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() {t.keyDown()} )
	}
}

func RightBindingOption(k KeyBinding) TreeViewOption {
	return func (t *TreeView) {
		displayBinding := KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   AlwaysFalse,
			Hide:       AlwaysTrue,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() {t.keyRight()} )
	}
}

func LeftBindingOption(k KeyBinding) TreeViewOption {
	return func (t *TreeView) {
		displayBinding := KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   AlwaysFalse,
			Hide:       AlwaysTrue,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() {t.keyLeft()} )
	}
}

func PageUpBindingOption(k KeyBinding) TreeViewOption {
	return func (t *TreeView) {
		displayBinding := KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   AlwaysFalse,
			Hide:       AlwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() {t.pageUp()} )
	}
}


func PageDownBindingOption(k KeyBinding) TreeViewOption {
	return func (t *TreeView) {
		displayBinding := KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   AlwaysFalse,
			Hide:       AlwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() {t.pageDown()} )
	}
}


func CollapseDirBindingOption(k KeyBinding) TreeViewOption {
	return func (t *TreeView) {
		displayBinding := KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   AlwaysFalse,
			Hide:       AlwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.collapseDir() } )
	}
}

func CollapseAllBindingOption(k KeyBinding) TreeViewOption {
	return func (t *TreeView) {
		displayBinding := KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   AlwaysFalse,
			Hide:       AlwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() { t.CollapseOrExpandAll() } )
	}
}

func ToggleAttributesOption(k KeyBinding) TreeViewOption {
	return func (t *TreeView) {
		displayBinding := KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   func() bool {return t.showAttributes},
			Hide:       AlwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func() {t.showAttributes = !t.showAttributes})
	}
}

func ToggleAddedFilesOption(k KeyBinding) TreeViewOption {
	return func (t *TreeView) {
		displayBinding := KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   func() bool {return t.tree.GetHiddenFileType(filetree.Added)},
			Hide:       AlwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func()  { t.tree.ToggleHiddenFileType(filetree.Added) })
	}
}

func ToggleRemovedFilesOption(k KeyBinding) TreeViewOption {
	return func (t *TreeView) {
		displayBinding := KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   func() bool {return t.tree.GetHiddenFileType(filetree.Removed)},
			Hide:       AlwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func()  { t.tree.ToggleHiddenFileType(filetree.Removed) })
	}
}

func ToggleModifiedFilesOption(k KeyBinding) TreeViewOption {
	return func (t *TreeView) {
		displayBinding := KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   func() bool {return t.tree.GetHiddenFileType(filetree.Modified)},
			Hide:       AlwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func()  { t.tree.ToggleHiddenFileType(filetree.Modified) })

	}
}


func ToggleUnmodifiedFilesOption(k KeyBinding) TreeViewOption {
	return func (t *TreeView) {
		displayBinding := KeyBindingDisplay{
			KeyBinding: &k,
			Selected:   func() bool {return t.tree.GetHiddenFileType(filetree.Unmodified)},
			Hide:       AlwaysFalse,
		}
		t.keyInputHandler.AddBinding(displayBinding, func()  { t.tree.ToggleHiddenFileType(filetree.Unmodified) })
	}
}


type TreeView struct {
	*tview.Box
	tree TreeModel

	// Note that the following two fields are distinct
	// treeIndex is the index about where we are in the current fileTree
	// this should be updated every keypress
	treeIndex int

	bufferIndexLowerBound int

	globalCollapseAll bool

	keyInputHandler *KeyInputHandler

	showAttributes bool
}

func NewTreeView(tree TreeModel) *TreeView {
	return &TreeView{
		Box:               tview.NewBox(),
		tree:              tree,
		globalCollapseAll: true,
		showAttributes:    true,
		keyInputHandler: NewKeyInputHandler(),
	}
}

func (t *TreeView) AddBindingOptions(bindingOptions ...TreeViewOption) *TreeView {
	for _, option := range bindingOptions {
		option(t)
	}

	return t
}

type KeyBindingConfig interface {
	GetKeyBinding(key string) (KeyBinding, error)
}


// Implementation notes:
 //need to set up our input handler here,
 //Should probably factor out keybinding initialization into a new function

func (t *TreeView) Setup(config KeyBindingConfig) *TreeView {
	t.tree.SetLayerIndex(0)

	t.AddBindingOptions(
		UpBindingOption(NewKeyBinding("Cursor Up", tcell.NewEventKey(tcell.KeyUp, rune(0), tcell.ModNone))),
		DownBindingOption(NewKeyBinding("Cursor Down", tcell.NewEventKey(tcell.KeyDown, rune(0), tcell.ModNone))),
		LeftBindingOption(NewKeyBinding("Cursor Left", tcell.NewEventKey(tcell.KeyLeft, rune(0), tcell.ModNone))),
		RightBindingOption(NewKeyBinding("Cursor Right", tcell.NewEventKey(tcell.KeyRight, rune(0), tcell.ModNone))),
	)

	bindingOrder := []string{
		"keybinding.toggle-collapse-dir",
		"keybinding.toggle-collapse-all-dir",
		"keybinding.toggle-filetree-attributes",
		"keybinding.toggle-added-files",
		"keybinding.toggle-removed-files",
		"keybinding.toggle-modified-files",
		"keybinding.toggle-unmodified-files",
		"keybinding.page-up",
		"keybinding.page-down",
	}

	bindingSettings := map[string]func(KeyBinding) TreeViewOption{
		"keybinding.toggle-collapse-dir":        CollapseDirBindingOption,
		"keybinding.toggle-collapse-all-dir":    CollapseAllBindingOption,
		"keybinding.toggle-filetree-attributes": ToggleAttributesOption,
		"keybinding.toggle-added-files":         ToggleAddedFilesOption,
		"keybinding.toggle-removed-files":       ToggleRemovedFilesOption,
		"keybinding.toggle-modified-files":      ToggleModifiedFilesOption,
		"keybinding.toggle-unmodified-files":    ToggleUnmodifiedFilesOption,
		"keybinding.page-up":                    PageUpBindingOption,
		"keybinding.page-down":                  PageDownBindingOption,
	}

	for _, keybinding := range bindingOrder {
		action := bindingSettings[keybinding]
		binding, err := config.GetKeyBinding(keybinding)
		if err != nil {
			panic(fmt.Errorf("setup error during %s: %w", keybinding, err))
		}

		t.AddBindingOptions(action(binding))
	}

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

func (t *TreeView) GetKeyBindings() []KeyBindingDisplay {
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
		logrus.Debugf("collapsing node %s", node.Path())
		node.Data.ViewInfo.Collapsed = !node.Data.ViewInfo.Collapsed
	}
	if node != nil {
		logrus.Debugf("unable to collapse node %s", node.Path())
		logrus.Debugf("  IsDir: %t", node.Data.FileInfo.IsDir)

	} else {
		logrus.Debugf("unable to collapse nil node")
	}

	return true
}

func (t *TreeView) CollapseOrExpandAll() bool {
	visitor := func(n *filetree.FileNode) error {
		if n.Data.FileInfo.IsDir {
			n.Data.ViewInfo.Collapsed = t.globalCollapseAll
		}
		return nil
	}

	evaluator := func(n *filetree.FileNode) bool {
		return true
	}
	if err := t.tree.VisitDepthParentFirst(visitor, evaluator); err != nil {
		panic(fmt.Errorf("error callapsing all dir: %w", err))
		// TODO log error here
		//return false
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
		logrus.Errorf("unable to get node position: %+v", err)
	}

	return node
}


func (t *TreeView) GetInnerRect() (int,int,int,int) {
	x, y, width, height := t.Box.GetInnerRect()
	return x, y+1, width, height-1
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

	logrus.Debugf("  treeIndex: %d", t.treeIndex)
	logrus.Debugf("  bufferIndexLowerBound: %d", t.bufferIndexLowerBound)
	logrus.Debugf("  height: %d", height)

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

	logrus.Debugf("keyUp end at: %s", t.getAbsPositionNode().Path())
	logrus.Debugf("  treeIndex: %d", t.treeIndex)
	logrus.Debugf("  bufferIndexLowerBound: %d", t.bufferIndexLowerBound)
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
