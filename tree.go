package main

import (
	"sort"
	"strings"
	"fmt"
	"errors"
)

const (
	newLine        = "\n"
	emptySpace     = "    "
	middleItem     = "├── "
	continueItem   = "│   "
	lastItem       = "└── "
	whiteoutPrefix = ".wh."
)

//type FileTree interface {
//	AddPath(string, interface{}) *Node
//	RemovePath(string) error
//	Visit(Visiter) error
//	// Diff(*FileTree) error
//	Stack(*FileTree) (FileTree, error)
//}


type FileTree struct {
	root *Node
	size int
	name string
}

type Node struct {
	tree     *FileTree
	parent   *Node
	name     string
	data     interface{}
	children map[string]*Node
}

func NewTree() (tree *FileTree) {
	tree = new(FileTree)
	tree.size = 0
	tree.root = new(Node)
	tree.root.tree = tree
	tree.root.children = make(map[string]*Node)
	return tree
}

func NewNode(parent *Node, name string, data interface{}) (node *Node) {
	node = new(Node)
	node.name = name
	node.data = data
	node.children = make(map[string]*Node)
	node.parent = parent
	node.tree = parent.tree
	return node
}

func (tree *FileTree) Root() *Node {
	return tree.root
}

func (node *Node) AddChild(name string, data interface{}) (child *Node) {
	child = NewNode(node, name, data)
	if node.children[name] != nil {
		// tree node already exists, replace the payload, keep the children
		node.children[name].data = data
	} else {
		node.children[name] = child
		node.tree.size++
	}
	return child
}

func (node *Node) Remove() error {
	for _, child := range node.children {
		child.Remove()
	}
	delete(node.parent.children, node.name)
	node.tree.size--
	return nil
}

func (node *Node) String() string {
	return node.name
}

func (tree *FileTree) String() string {
	var renderLine func(string, []bool, bool) string
	var walkTree func(*Node, []bool) string

	renderLine = func(text string, spaces []bool, last bool) string {
		var result string
		for _, space := range spaces {
			if space {
				result += emptySpace
			} else {
				result += continueItem
			}
		}

		indicator := middleItem
		if last {
			indicator = lastItem
		}

		return result + indicator + text + newLine
	}

	walkTree = func(node *Node, spaces []bool) string {
		var result string
		var keys []string
		for key := range node.children {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for idx, name := range keys {
			child := node.children[name]
			last := idx == (len(node.children) - 1)
			result += renderLine(child.String(), spaces, last)
			if len(child.children) > 0 {
				spacesChild := append(spaces, last)
				result += walkTree(child, spacesChild)
			}
		}
		return result
	}

	return "." + newLine + walkTree(tree.Root(), []bool{})
}

func (node *Node) Copy() *Node {
	// newNode := new(Node)
	// *newNode = *node
	// return newNode
	newNode := NewNode(node.parent, node.name, node.data)
	for name, child := range node.children {
		newNode.children[name] = child.Copy()
	}
	return newNode
}

func (tree *FileTree) Copy() *FileTree {
	newTree := NewTree()
	*newTree = *tree
	newTree.root = tree.Root().Copy()

	return newTree
}

type Visiter func(*Node) error

func (tree *FileTree) Visit(visiter Visiter) error {
	return tree.root.Visit(visiter)
}

func (node *Node) Visit(visiter Visiter) error {
	for _, child := range node.children {
		err := child.Visit(visiter)
		if err != nil {
			return err
		}
	}
	return visiter(node)
}

func (node *Node) IsWhiteout() bool {
	return strings.HasPrefix(node.name, whiteoutPrefix)
}

func (node *Node) Path() string {
	path := []string{}
	curNode := node
	for {
		if curNode.parent == nil{
			break
		}

		name := curNode.name
		if curNode == node {
			// white out prefixes are fictitious on leaf nodes
			name = strings.TrimPrefix(name, whiteoutPrefix)
		}

		path = append([]string{name}, path...)
		curNode = curNode.parent
	}
	return "/" + strings.Join(path, "/")
}



func (tree *FileTree) Stack(upper *FileTree) (error) {
	graft := func(node *Node) error {
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

func (tree *FileTree) GetNode(path string) (*Node, error) {
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

func (tree *FileTree) AddPath(path string, data interface{}) (*Node, error) {
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


