package components

import (
	"bytes"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
	"io"
	"strings"
)

// TODO simplify this interface.
type TreeModel interface {
	StringBetween(int, int, bool) string
	VisitDepthParentFirst(filetree.Visitor, filetree.VisitEvaluator) error
	VisitDepthChildFirst(filetree.Visitor, filetree.VisitEvaluator) error
	RemovePath(path string) error
	VisibleSize() int
	SetLayerIndex(int) bool
}

type TreeView struct {
	*tview.Box
	tree TreeModel

	// Note that the following two fields are distinct
	// treeIndex is the index about where we are in the current fileTree
	// this should be updated every keypress
	treeIndex int

	bufferIndexLowerBound int
}

func NewTreeView(tree TreeModel) *TreeView {
	return &TreeView{
		Box: tview.NewBox(),
		tree: tree,
	}
}

func (t *TreeView) Setup() *TreeView {
	t.SetBorder(true).
		SetTitle("Files").
		SetTitleAlign(tview.AlignLeft)
	t.tree.SetLayerIndex(0)

	return t
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
	})
}

func (t *TreeView) Focus(delegate func(p tview.Primitive)) {
	t.Box.Focus(delegate)
}

func (t *TreeView) HasFocus() bool {
	return t.Box.HasFocus()
}
// Private helper methods

func (t *TreeView) spaceDown() bool {
	node := t.getAbsPositionNode()
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

func (t *TreeView) keyDown() bool {
	_, _, _, height := t.Box.GetInnerRect()

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

func (t *TreeView) bufferIndexUpperBound() int {
	_,_, _, height := t.Box.GetInnerRect()
	return t.bufferIndexLowerBound + height
}


func (t *TreeView) Draw(screen tcell.Screen) {
	t.Box.Draw(screen)
	selectedIndex := t.treeIndex - t.bufferIndexLowerBound
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
			style = style.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack).Bold(true)
			if yIndex == selectedIndex {
				screen.SetContent(x+xIndex, y+yIndex, m, c, style)
				screen.SetContent(x+xIndex, y+yIndex, m, c, style)
			} else if yIndex > selectedIndex {
				break
			}
		}
	}

}
