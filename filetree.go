package main

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
	root *FileNode
	size int
	name string
}

func NewTree() (tree *FileTree) {
	tree = new(FileTree)
	tree.size = 0
	tree.root = new(FileNode)
	tree.root.tree = tree
	tree.root.children = make(map[string]*FileNode)
	return tree
}

func (tree *FileTree) Root() *FileNode {
	return tree.root
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
		for key := range node.children {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for idx, name := range keys {
			child := node.children[name]
			last := idx == (len(node.children) - 1)
			showCollapsed := child.collapsed && len(child.children) > 0
			result += renderLine(child.String(), spaces, last, showCollapsed)
			if len(child.children) > 0 && !child.collapsed {
				spacesChild := append(spaces, last)
				result += walkTree(child, spacesChild, depth+1)
			}
		}
		return result
	}

	return "." + newLine + walkTree(tree.Root(), []bool{}, 0)
}

func (tree *FileTree) Copy() *FileTree {
	newTree := NewTree()
	*newTree = *tree
	newTree.root = tree.Root().Copy()
	newTree.Visit(func(node *FileNode) error {
		node.tree = newTree
		return nil
	})

	return newTree
}

type Visiter func(*FileNode) error
type VisitEvaluator func(*FileNode) bool

func (tree *FileTree) Visit(visiter Visiter) error {
	return tree.root.Visit(visiter)
}

func (tree *FileTree) VisitDepthParentFirst(visiter Visiter, evaluator VisitEvaluator) error {
	return tree.root.VisitDepthParentFirst(visiter, evaluator)
}

func (tree *FileTree) Stack(upper *FileTree) error {
	graft := func(node *FileNode) error {
		if node.IsWhiteout() {
			err := tree.RemovePath(node.Path())
			if err != nil {
				return fmt.Errorf("Cannot remove node %s: %v", node.Path(), err.Error())
			}
		} else {
			newNode, err := tree.AddPath(node.Path(), node.data)
			if err != nil {
				return fmt.Errorf("Cannot add node %s: %v", newNode.Path(), err.Error())
			}
		}
		return nil
	}
	return upper.Visit(graft)
}

func (tree *FileTree) GetNode(path string) (*FileNode, error) {
	nodeNames := strings.Split(path, "/")
	node := tree.Root()
	for _, name := range nodeNames {
		if name == "" {
			continue
		}
		if node.children[name] == nil {
			return nil, errors.New("Path does not exist")
		}
		node = node.children[name]
	}
	return node, nil
}

func (tree *FileTree) AddPath(path string, data *FileChangeInfo) (*FileNode, error) {
	nodeNames := strings.Split(path, "/")
	node := tree.Root()
	for idx, name := range nodeNames {
		if name == "" {
			continue
		}
		// find or create node
		if node.children[name] != nil {
			node = node.children[name]
		} else {
			// don't attach the payload. The payload is destined for the
			// path's end node, not any intermediary node.
			node = node.AddChild(name, nil)
		}

		// attach payload to the last specified node
		if idx == len(nodeNames)-1 {
			node.data = data
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

func (tree *FileTree) compare(upper *FileTree) error {
	graft := func(node *FileNode) error {
		if node.IsWhiteout() {
			err := tree.MarkRemoved(node.Path())
			if err != nil {
				return fmt.Errorf("Cannot remove node %s: %v", node.Path(), err.Error())
			}
		} else {
			existingNode, _ := tree.GetNode(node.Path())
			if existingNode == nil {
				newNode, err := tree.AddPath(node.Path(), node.data)
				fmt.Printf("added new node at %s\n", newNode.Path())
				if err != nil {
					return fmt.Errorf("Cannot add new node %s: %v", node.Path(), err.Error())
				}
				newNode.AssignDiffType(Added)
			} else {
				diffType := existingNode.compare(node)
				fmt.Printf("found existing node at %s\n", existingNode.Path())
				existingNode.deriveDiffType(diffType)
			}
		}
		return nil
	}
	return upper.Visit(graft)
}

func (tree *FileTree) MarkRemoved(path string) error {
	node, err := tree.GetNode(path)
	if err != nil {
		return err
	}
	return node.AssignDiffType(Removed)
}

func StackRange(trees []*FileTree, index uint) *FileTree {
	tree := trees[1].Copy()
	for idx := uint(2); idx < index; idx++ {
		tree.Stack(trees[idx])
	}
	return tree
}
