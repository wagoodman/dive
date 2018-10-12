package filetree

import (
	"fmt"
	"strings"
	"github.com/satori/go.uuid"
	"sort"
)

const (
	newLine         = "\n"
	noBranchSpace   = "    "
	branchSpace     = "│   "
	middleItem      = "├─"
	lastItem        = "└─"
	whiteoutPrefix  = ".wh."
	uncollapsedItem = "─ "
	collapsedItem   = "⊕ "
)

type FileTree struct {
	Root *FileNode
	Size int
	FileSize uint64
	Name string
	Id   uuid.UUID
}

func NewFileTree() (tree *FileTree) {
	tree = new(FileTree)
	tree.Size = 0
	tree.Root = new(FileNode)
	tree.Root.Tree = tree
	tree.Root.Children = make(map[string]*FileNode)
	tree.Id = uuid.Must(uuid.NewV4())
	return tree
}

type renderParams struct{
	node *FileNode
	spaces []bool
	childSpaces []bool
	showCollapsed bool
	isLast bool
}

func (tree *FileTree) renderStringTreeBetween(startRow, stopRow int, showAttributes bool) string {
	// generate a list of nodes to render
	var params []renderParams = make([]renderParams,0)
	var result string

	// visit from the front of the list
	var paramsToVisit = []renderParams{ renderParams{node: tree.Root, spaces: []bool{}, showCollapsed: false, isLast: false} }
	for currentRow := 0; len(paramsToVisit) > 0 && currentRow <= stopRow; currentRow++ {
		// pop the first node
		var currentParams renderParams
		currentParams, paramsToVisit = paramsToVisit[0], paramsToVisit[1:]

		// take note of the next nodes to visit later
		var keys []string
		for key := range currentParams.node.Children {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		var childParams = make([]renderParams,0)
		for idx, name := range keys {
			child := currentParams.node.Children[name]
			// don't visit this node...
			if child.Data.ViewInfo.Hidden || currentParams.node.Data.ViewInfo.Collapsed {
				continue
			}

			// visit this node...
			isLast := idx == (len(currentParams.node.Children) - 1)
			showCollapsed := child.Data.ViewInfo.Collapsed && len(child.Children) > 0

			// completely copy the reference slice
			childSpaces := make([]bool, len(currentParams.childSpaces))
			copy(childSpaces, currentParams.childSpaces)

			if len(child.Children) > 0 && !child.Data.ViewInfo.Collapsed {
				childSpaces = append(childSpaces, isLast)
			}

			childParams = append(childParams, renderParams{
				node: child,
				spaces: currentParams.childSpaces,
				childSpaces: childSpaces,
				showCollapsed: showCollapsed,
				isLast: isLast,
			})
		}
		// keep the child nodes to visit later
		paramsToVisit = append(childParams, paramsToVisit...)

		// never process the root node
		if currentParams.node == tree.Root {
			currentRow--
			continue
		}

		// process the current node
		if currentRow >= startRow && currentRow <= stopRow {
			params = append(params, currentParams)
		}
	}

	// render the result
	for idx := range params {
		currentParams := params[idx]

		if showAttributes {
			result += currentParams.node.MetadataString() + " "
		}
		result += currentParams.node.renderTreeLine(currentParams.spaces, currentParams.isLast, currentParams.showCollapsed)
	}

	return result
}

func (tree *FileTree) String(showAttributes bool) string {
	return tree.renderStringTreeBetween(0, tree.Size, showAttributes)
}

func (tree *FileTree) StringBetween(start, stop uint, showAttributes bool) string {
	return tree.renderStringTreeBetween(int(start), int(stop), showAttributes)
}

func (tree *FileTree) Copy() *FileTree {
	newTree := NewFileTree()
	newTree.Size = tree.Size
	newTree.FileSize = tree.FileSize
	newTree.Root = tree.Root.Copy(newTree.Root)

	// update the tree pointers
	newTree.VisitDepthChildFirst(func(node *FileNode) error {
		node.Tree = newTree
		return nil
	}, nil)

	return newTree
}

type Visiter func(*FileNode) error
type VisitEvaluator func(*FileNode) bool

// DFS bubble up
func (tree *FileTree) VisitDepthChildFirst(visiter Visiter, evaluator VisitEvaluator) error {
	return tree.Root.VisitDepthChildFirst(visiter, evaluator)
}

// DFS sink down
func (tree *FileTree) VisitDepthParentFirst(visiter Visiter, evaluator VisitEvaluator) error {
	return tree.Root.VisitDepthParentFirst(visiter, evaluator)
}

func (tree *FileTree) Stack(upper *FileTree) error {
	graft := func(node *FileNode) error {
		if node.IsWhiteout() {
			err := tree.RemovePath(node.Path())
			if err != nil {
				return fmt.Errorf("cannot remove node %s: %v", node.Path(), err.Error())
			}
		} else {
			newNode, err := tree.AddPath(node.Path(), node.Data.FileInfo)
			if err != nil {
				return fmt.Errorf("cannot add node %s: %v", newNode.Path(), err.Error())
			}
		}
		return nil
	}
	return upper.VisitDepthChildFirst(graft, nil)
}

func (tree *FileTree) GetNode(path string) (*FileNode, error) {
	nodeNames := strings.Split(strings.Trim(path, "/"), "/")
	node := tree.Root
	for _, name := range nodeNames {
		if name == "" {
			continue
		}
		if node.Children[name] == nil {
			return nil, fmt.Errorf("path does not exist: %s", path)
		}
		node = node.Children[name]
	}
	return node, nil
}

func (tree *FileTree) AddPath(path string, data FileInfo) (*FileNode, error) {
	nodeNames := strings.Split(strings.Trim(path, "/"), "/")
	node := tree.Root
	for idx, name := range nodeNames {
		if name == "" {
			continue
		}
		// find or create node
		if node.Children[name] != nil {
			node = node.Children[name]
		} else {
			// don't attach the payload. The payload is destined for the
			// Path's end node, not any intermediary node.
			node = node.AddChild(name, FileInfo{})
		}

		// attach payload to the last specified node
		if idx == len(nodeNames)-1 {
			node.Data.FileInfo = data
		}

	}
	return node, nil
}

func (tree *FileTree) RemovePath(path string) error {
	node, err := tree.GetNode(path)
	if err != nil {
		return err
	}
	return node.Remove()
}

func (tree *FileTree) Compare(upper *FileTree) error {
	graft := func(upperNode *FileNode) error {
		if upperNode.IsWhiteout() {
			err := tree.MarkRemoved(upperNode.Path())
			if err != nil {
				return fmt.Errorf("cannot remove upperNode %s: %v", upperNode.Path(), err.Error())
			}
		} else {
			lowerNode, _ := tree.GetNode(upperNode.Path())
			if lowerNode == nil {
				newNode, err := tree.AddPath(upperNode.Path(), upperNode.Data.FileInfo)
				if err != nil {
					return fmt.Errorf("cannot add new upperNode %s: %v", upperNode.Path(), err.Error())
				}
				newNode.AssignDiffType(Added)
			} else {
				diffType := lowerNode.compare(upperNode)
				return lowerNode.deriveDiffType(diffType)
			}
		}
		return nil
	}
	return upper.VisitDepthChildFirst(graft, nil)
}

func (tree *FileTree) MarkRemoved(path string) error {
	node, err := tree.GetNode(path)
	if err != nil {
		return err
	}
	return node.AssignDiffType(Removed)
}

func StackRange(trees []*FileTree, start, stop int) *FileTree {
	tree := trees[0].Copy()
	for idx := start; idx <= stop; idx++ {
		tree.Stack(trees[idx])
	}

	return tree
}
