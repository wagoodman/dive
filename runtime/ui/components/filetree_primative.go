package components

import (
	"bytes"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
	"io"
	"regexp"
	"strings"
)

// TODO simplify this interface.
type TreeModel interface {
	StringBetween(int, int, bool) string
	VisitDepthParentFirst(filetree.Visitor, filetree.VisitEvaluator) error
	VisitDepthChildFirst(filetree.Visitor, filetree.VisitEvaluator) error
	RemovePath(path string) error
	VisibleSize() int
}

type TreeView struct {
	*tview.Box
	tree TreeModel

	// Note that the following two fields are distinct
	// treeIndex is the index about where we are in the current fileTree
	// this should be updated every keypresws
	treeIndex int

	// bufferIndex is the index about where we are in the Buffer,
	// basically lets us scroll down but NOT shift the buffer
	bufferIndexLowerBound int
	bufferIndex int

	filterRegex *regexp.Regexp
	//changed func(index int, mainText string, shortcut rune)

}

func NewTreeView(tree TreeModel) *TreeView {
	return &TreeView{
		Box: tview.NewBox(),
		tree: tree,
	}
}

// TODO: make these keys configurable
func (t *TreeView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp:
			t.keyUp()
		case tcell.KeyDown:
			t.keyDown()
		case tcell.KeyRight:
			t.keyRight()
		case tcell.KeyLeft:
			t.keyLeft()
		}
		switch event.Rune() {
		case ' ':
			t.spaceDown()
		}
		//t.changed(t.cmpIndex, t.layers[t.cmpIndex], event.Rune())
	})
}

func (t *TreeView) SetTree(newTree TreeModel) *TreeView {
	// preserve collapsed nodes based on path
	collapsedList := map[string]interface{}{}

	evaluateFunc := func(node *filetree.FileNode) bool {
		if node.Parent != nil && (node.Parent.Data.ViewInfo.Collapsed || node.Parent.Data.ViewInfo.Hidden) {
			return false
		}
		return true
	}

	t.tree.VisitDepthParentFirst(func(node *filetree.FileNode) error {
		if node.Data.ViewInfo.Collapsed {
			collapsedList[node.Path()] = true
		}
		return nil
	},evaluateFunc)

	newTree.VisitDepthParentFirst(func(node *filetree.FileNode) error {
		_, ok := collapsedList[node.Path()]
		if ok {
			node.Data.ViewInfo.Collapsed = true
		}
		return nil
	}, evaluateFunc)

	t.tree = newTree
	if err := t.FilterUpdate(); err != nil {
		panic(err)
	}

	return t
}

func (t *TreeView) GetTree() TreeModel {
	return t.tree
}

func (t *TreeView) Focus(delegate func(p tview.Primitive)) {
	t.Box.Focus(delegate)
}

func (t *TreeView) HasFocus() bool {
	return t.Box.HasFocus()
}

func (t *TreeView) SetFilterRegex(filterRegex *regexp.Regexp) {
	t.filterRegex = filterRegex
	if err := t.FilterUpdate(); err != nil {
		panic(err)
	}
}

// Private helper methods

func (t *TreeView) spaceDown() bool {
	node := t.getAbsPositionNode(nil)
	if node != nil && node.Data.FileInfo.IsDir {
		logrus.Debugf("collapsing node %s", node.Path())
		node.Data.ViewInfo.Collapsed = !node.Data.ViewInfo.Collapsed
		return true
	}
	if node != nil {
		logrus.Debugf("unable to collapse node %s", node.Path())
		logrus.Debugf("  IsDir: %t", node.Data.FileInfo.IsDir)

	} else {
		logrus.Debugf("unable to collapse nil node")
	}
	return false
}

// getAbsPositionNode determines the selected screen cursor's location in the file tree, returning the selected FileNode.
func (t *TreeView) getAbsPositionNode(filterRegex *regexp.Regexp) (node *filetree.FileNode) {
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
		regexMatch := true
		if filterRegex != nil {
			match := filterRegex.Find([]byte(curNode.Path()))
			regexMatch = match != nil
		}
		return !curNode.Parent.Data.ViewInfo.Collapsed && !curNode.Data.ViewInfo.Hidden && regexMatch
	}

	err := t.tree.VisitDepthParentFirst(visitor, evaluator)
	if err != nil {
		logrus.Errorf("unable to get node position: %+v", err)
	}

	return node
}

func (t *TreeView) keyDown() bool {
	_, _, _, height := t.Box.GetInnerRect()

	// treeIndex is the index about where we are in the current file
	if t.treeIndex >= t.tree.VisibleSize() {
		return false
	}
	t.treeIndex++
	if t.treeIndex > height {
		t.bufferIndexLowerBound++
	}
	t.bufferIndex++
	if t.bufferIndex > height {
		t.bufferIndex = height
	}
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
	if t.bufferIndex > 0 {
		t.bufferIndex--
	}
	return true
}

// TODO add regex filtering
func (t *TreeView) keyRight() bool {
	node := t.getAbsPositionNode(t.filterRegex)

	_,_, _, height := t.Box.GetInnerRect()
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
	if t.treeIndex > t.bufferIndexUpperBound() {
		t.bufferIndexLowerBound++
	}

	t.bufferIndex++
	if t.bufferIndex > height {
		t.bufferIndex = height
	}

	return true
}

func (t *TreeView) keyLeft() bool {
	var visitor func(*filetree.FileNode) error
	var evaluator func(*filetree.FileNode) bool
	var dfsCounter, newIndex int
	oldIndex := t.treeIndex
	currentNode := t.getAbsPositionNode(t.filterRegex)

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
		regexMatch := true
		if t.filterRegex != nil {
			match := t.filterRegex.Find([]byte(curNode.Path()))
			regexMatch = match != nil
		}
		return !curNode.Parent.Data.ViewInfo.Collapsed && !curNode.Data.ViewInfo.Hidden && regexMatch
	}

	err := t.tree.VisitDepthParentFirst(visitor, evaluator)
	if err != nil {
		// TODO: remove this panic
		panic(err)
	}

	t.treeIndex = newIndex
	moveIndex := oldIndex - newIndex
	if newIndex < t.bufferIndexLowerBound {
		t.bufferIndexLowerBound = t.treeIndex
	}

	if t.bufferIndex > moveIndex {
		t.bufferIndex -= moveIndex
	} else {
		t.bufferIndex = 0
	}

	return true
}

func (t *TreeView) bufferIndexUpperBound() int {
	_,_, _, height := t.Box.GetInnerRect()
	return t.bufferIndexLowerBound + height
}

func (t *TreeView) FilterUpdate() error {
	// keep the t selection in parity with the current DiffType selection
	err := t.tree.VisitDepthChildFirst(func(node *filetree.FileNode) error {
		// TODO: add hidden datatypes.
		//node.Data.ViewInfo.Hidden = t.HiddenDiffTypes[node.Data.DiffType]
		visibleChild := false
		if t.filterRegex == nil {
			node.Data.ViewInfo.Hidden = false
			return nil
		}

		for _, child := range node.Children {
			if !child.Data.ViewInfo.Hidden {
				visibleChild = true
				node.Data.ViewInfo.Hidden = false
				return nil
			}
		}

		if !visibleChild { // hide nodes that do not match the current file filter regex (also don't unhide nodes that are already hidden)
			match := t.filterRegex.FindString(node.Path())
			node.Data.ViewInfo.Hidden = len(match) == 0
		}
		return nil
	}, nil)

	if err != nil {
		logrus.Errorf("unable to propagate t model tree: %+v", err)
		return err
	}

	return nil
}


func (t *TreeView) Draw(screen tcell.Screen) {
	t.Box.Draw(screen)

	x, y, width, height := t.Box.GetInnerRect()
	showAttributes := width > 80
	// TODO add switch for showing attributes.
	treeString := t.tree.StringBetween(t.bufferIndexLowerBound, t.bufferIndexUpperBound(), showAttributes)
	lines := strings.Split(treeString, "\n")

	// update the contents
	for yIndex, line := range lines {
		if yIndex >= height {
			break
		}
		// Strip out ansi colors, Tview cannot use these
		stripLine := bytes.NewBuffer(nil)
		w := tview.ANSIWriter(stripLine)
		if _, err := io.Copy(w, strings.NewReader(line)); err != nil  {
			//TODO: handle panic gracefully
			panic(err)
		}

		tview.Print(screen, stripLine.String(), x, y+yIndex, width, tview.AlignLeft, tcell.ColorDefault)
		for xIndex := 0; xIndex < width; xIndex++ {
			m, c, style, _ := screen.GetContent(x+xIndex, y+yIndex)
			// TODO make these background an forground colors flexable
			style = style.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack).Bold(true)
			if yIndex == t.bufferIndex {
				screen.SetContent(x+xIndex, y+yIndex, m, c, style)
				screen.SetContent(x+xIndex, y+yIndex, m, c, style)
			} else if yIndex > t.bufferIndex {
				break
			}
		}
	}

}
