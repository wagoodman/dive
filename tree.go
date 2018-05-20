package main

import "errors"

type Entity interface {
	Visit(Visiter)
}

type Tree struct {
	root Node
	size uint
}

type Node struct {
	tree     *Tree
	parent   *Node
	name     string
	data     interface{}
	children map[string]*Node
}

type Visiter func(*Node)

func NewTree() (tree *Tree) {
	tree = new(Tree)
	tree.size = 1
	root := &tree.root
	root.tree = tree
	root.children = make(map[string]*Node)
	return tree
}

func NewNode(name string, data interface{}) (node *Node) {
	node.name = name
	node.data = data
	node.children = make(map[string]*Node)
	return node
}

func (tree *Tree) Root() *Node {
	return &tree.root
}

func (tree *Tree) Visit(visiter Visiter) {
	tree.root.Visit(visiter)
}

func (parent *Node) Add(name string, data interface{}) (child *Node, error error) {
	if parent.children[name] != nil {
		return nil, errors.New("Duplicate child")
	}
	child = NewNode(name, data)
	child.tree = parent.tree
	child.parent = parent
	child.tree.size++
	parent.children[name] = child
	return child, nil
}

func (node *Node) Visit(visiter Visiter) {
	for _, child := range node.children {
		child.Visit(visiter)
	}
	visiter(node)
}

func main() {

}
