package main

import (
	"sort"
)

const (
	newLine      = "\n"
	emptySpace   = "    "
	middleItem   = "├── "
	continueItem = "│   "
	lastItem     = "└── "
)

type Tree struct {
	root *Node
	size int
}

type Node struct {
	tree     *Tree
	parent   *Node
	name     string
	data     interface{}
	children map[string]*Node
}

func NewTree() (tree *Tree) {
	tree = new(Tree)
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

func (tree *Tree) Root() *Node {
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

func (tree *Tree) String() string {
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

func (tree *Tree) Copy() *Tree {
	newTree := NewTree()
	*newTree = *tree
	newTree.root = tree.Root().Copy()

	return newTree
}

