package main

import (
	"errors"
	"strings"
)

type FileTree interface {
	AddPath(string, interface{})
	RemovePath(string)
	Visit(Visiter)
}

type Visiter func(*Node)

func (tree *Tree) Visit(visiter Visiter) {
	tree.root.Visit(visiter)
}

func (node *Node) Visit(visiter Visiter) {
	for _, child := range node.children {
		child.Visit(visiter)
	}
	visiter(node)
}

func (tree *Tree) AddPath(path string, data interface{}) (*Node, error) {
	nodeNames := strings.Split(path, "/")
	node := tree.Root()
	var err error
	for idx, name := range nodeNames {
		if name == "" {
			continue
		}
		// find or create node
		if node.children[name] != nil {
			node = node.children[name]
		} else {

			node, _ = node.AddChild(name, nil)
			if err != nil {

				return node, err
			}
		}

		// attach payload
		if idx == len(nodeNames)-1 {
			node.data = data
		}

	}
	return node, nil
}

func (tree *Tree) RemovePath(path string) error {
	nodeNames := strings.Split(path, "/")
	node := tree.Root()
	for _, name := range nodeNames {
		if name == "" {
			continue
		}
		if node.children[name] == nil {
			return errors.New("Path does not exist")
		}
		node = node.children[name]
	}
	// this node's parent should be a leaf
	return node.Remove()
}
