package main

import (
	"errors"
	"strings"
	"fmt"
)

type FileTree interface {
	AddPath(string, interface{}) *Node
	RemovePath(string) error
	Visit(Visiter) error
	// Diff(*Tree) error
	Stack(*Tree) (Tree, error)
}

type Visiter func(*Node) error

func (tree *Tree) Visit(visiter Visiter) error {
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
	return strings.HasPrefix(node.name, ".wh.")
}

func (node *Node) Path() string {
	path := []string{}
	curNode := node
	for {
		if curNode.parent == nil{
			break
		}
		path = append([]string{curNode.name}, path...)
		curNode = curNode.parent
	}
	return "/" + strings.Join(path, "/")
}


func (node *Node) WhiteoutPath() string {
	path := []string{}
	curNode := node
	for {
		if curNode.parent == nil{
			break
		}

		name := curNode.name
		if curNode == node {
			name = strings.TrimPrefix(name, ".wh.")
		}

		path = append([]string{name}, path...)
		curNode = curNode.parent
	}
	return "/" + strings.Join(path, "/")
}



func (tree *Tree) Stack(upper *Tree) (error) {
	graft := func(node *Node) error {
		if node.IsWhiteout() {
			err := tree.RemovePath(node.WhiteoutPath())
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

func (tree *Tree) GetNode(path string) (*Node, error) {
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

func (tree *Tree) AddPath(path string, data interface{}) (*Node, error) {
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

func (tree *Tree) RemovePath(path string) error {
	node, err := tree.GetNode(path)
	if err != nil {
		return err
	}
	return node.Remove()
}
