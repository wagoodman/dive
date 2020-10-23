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

type TreeView struct {
	*tview.Box
	// TODO: make me an interface

	tree *filetree.FileTree


	// Note that the following two fields are distinct
	// treeIndex is the index about where we are in the current fileTree
	// this should be updated every keypresws
	treeIndex int

	// bufferIndex is the index about where we are in the Buffer,
	// basically lets us scroll down but NOT shift the buffer
	bufferIndexLowerBound int
	bufferIndex int

	//changed func(index int, mainText string, shortcut rune)
}

func NewTreeView(tree *filetree.FileTree) *TreeView {
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
		}
		switch event.Rune() {
		case ' ':
			t.spaceDown()
		}
		//t.changed(t.cmpIndex, t.layers[t.cmpIndex], event.Rune())
	})
}

func (t *TreeView) SetTree(newTree *filetree.FileTree) *TreeView {
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

	return t
}

func (t *TreeView) GetTree(tree *filetree.FileTree) *filetree.FileTree {
	return t.tree
}

func (t *TreeView) Focus(delegate func(p tview.Primitive)) {
	t.Box.Focus(delegate)
}

func (t *TreeView) HasFocus() bool {
	return t.Box.HasFocus()
}


// Private helper methods

func (t *TreeView) spaceDown() bool {
	node := t.getAbsPositionNode(nil)
	if node != nil && node.Data.FileInfo.IsDir {
		node.Data.ViewInfo.Collapsed = !node.Data.ViewInfo.Collapsed
		return true
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

func (t *TreeView) bufferIndexUpperBound() int {
	_,_, _, height := t.Box.GetInnerRect()
	return t.bufferIndexLowerBound + height

}

func (t *TreeView) Draw(screen tcell.Screen) {
	t.Box.Draw(screen)

	x, y, width, height := t.Box.GetInnerRect()

	// TODO add switch for showing attributes.
	treeString := t.tree.StringBetween(t.bufferIndexLowerBound, t.bufferIndexUpperBound(), false)
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
