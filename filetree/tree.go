package filetree

import (
	"errors"
	"fmt"
	"sort"
	"strings"
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
	Name string
}

func NewFileTree() (tree *FileTree) {
	tree = new(FileTree)
	tree.Size = 0
	tree.Root = new(FileNode)
	tree.Root.Tree = tree
	tree.Root.Children = make(map[string]*FileNode)
	return tree
}

func (tree *FileTree) String() string {
	var renderLine func(string, []bool, bool, bool) string
	var walkTree func(*FileNode, []bool, int) string

	renderLine = func(nodeText string, spaces []bool, last bool, collapsed bool) string {
		var otherBranches string
		for _, space := range spaces {
			if space {
				otherBranches += noBranchSpace
			} else {
				otherBranches += branchSpace
			}
		}

		thisBranch := middleItem
		if last {
			thisBranch = lastItem
		}

		collapsedIndicator := uncollapsedItem
		if collapsed {
			collapsedIndicator = collapsedItem
		}

		return otherBranches + thisBranch + collapsedIndicator + nodeText + newLine
	}

	walkTree = func(node *FileNode, spaces []bool, depth int) string {
		var result string
		var keys []string
		for key := range node.Children {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for idx, name := range keys {
			child := node.Children[name]
			if child.Data.ViewInfo.Hidden {
				continue
			}
			last := idx == (len(node.Children) - 1)
			showCollapsed := child.Data.ViewInfo.Collapsed && len(child.Children) > 0
			result += renderLine(child.String(), spaces, last, showCollapsed)
			if len(child.Children) > 0 && !child.Data.ViewInfo.Collapsed {
				spacesChild := append(spaces, last)
				result += walkTree(child, spacesChild, depth+1)
			}
		}
		return result
	}

	return "." + newLine + walkTree(tree.Root, []bool{}, 0)
}

func (tree *FileTree) Copy() *FileTree {
	newTree := NewFileTree()
	*newTree = *tree
	newTree.Root = tree.Root.Copy()
	newTree.VisitDepthChildFirst(func(node *FileNode) error {
		node.Tree = newTree
		return nil
	})

	return newTree
}

type Visiter func(*FileNode) error
type VisitEvaluator func(*FileNode) bool

// DFS bubble up
func (tree *FileTree) VisitDepthChildFirst(visiter Visiter) error {
	return tree.Root.VisitDepthChildFirst(visiter)
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
				return fmt.Errorf("Cannot remove node %s: %v", node.Path(), err.Error())
			}
		} else {
			newNode, err := tree.AddPath(node.Path(), node.Data.FileInfo)
			if err != nil {
				return fmt.Errorf("Cannot add node %s: %v", newNode.Path(), err.Error())
			}
		}
		return nil
	}
	return upper.VisitDepthChildFirst(graft)
}

func (tree *FileTree) GetNode(path string) (*FileNode, error) {
	nodeNames := strings.Split(path, "/")
	node := tree.Root
	for _, name := range nodeNames {
		if name == "" {
			continue
		}
		if node.Children[name] == nil {
			return nil, errors.New("Path does not exist")
		}
		node = node.Children[name]
	}
	return node, nil
}

func (tree *FileTree) AddPath(path string, data *FileInfo) (*FileNode, error) {
	// fmt.Printf("ADDPATH: %s %+v\n", path, data)
	nodeNames := strings.Split(path, "/")
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
			node = node.AddChild(name, nil)
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
	graft := func(node *FileNode) error {
		if node.IsWhiteout() {
			err := tree.MarkRemoved(node.Path())
			if err != nil {
				return fmt.Errorf("Cannot remove node %s: %v", node.Path(), err.Error())
			}
		} else {
			existingNode, _ := tree.GetNode(node.Path())
			if existingNode == nil {
				newNode, err := tree.AddPath(node.Path(), node.Data.FileInfo)
				// fmt.Printf("added new node at %s\n", newNode.Path())
				if err != nil {
					return fmt.Errorf("Cannot add new node %s: %v", node.Path(), err.Error())
				}
				newNode.AssignDiffType(Added)
			} else {
				diffType := existingNode.compare(node)
				// fmt.Printf("found existing node at %s\n", existingNode.Path())
				return existingNode.deriveDiffType(diffType)
			}
		}
		return nil
	}
	return upper.VisitDepthChildFirst(graft)
}

func (tree *FileTree) MarkRemoved(path string) error {
	node, err := tree.GetNode(path)
	if err != nil {
		return err
	}
	return node.AssignDiffType(Removed)
}

func StackRange(trees []*FileTree, index int) *FileTree {
	tree := trees[0].Copy()
	for idx := 0; idx <= index; idx++ {
		tree.Stack(trees[idx])
	}
	return tree
}
