package main

import (
	"sort"
	"strings"
)

type FileNode struct {
	tree      *FileTree
	parent    *FileNode
	name      string
	collapsed bool
	data       *FileChangeInfo
	children  map[string]*FileNode
}

func NewNode(parent *FileNode, name string, data *FileChangeInfo) (node *FileNode) {
	node = new(FileNode)
	node.name = name
	if data == nil {
		data = &FileChangeInfo{}
	}
	node.data = data
	node.children = make(map[string]*FileNode)
	node.parent = parent
	if parent != nil {
		node.tree = parent.tree
	}
	return node
}

func (node *FileNode) Copy() *FileNode {
	// newNode := new(FileNode)
	// *newNode = *node
	// return newNode
	newNode := NewNode(node.parent, node.name, node.data)
	for name, child := range node.children {
		newNode.children[name] = child.Copy()
	}
	return newNode
}


func (node *FileNode) AddChild(name string, data *FileChangeInfo) (child *FileNode) {
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

func (node *FileNode) Remove() error {
	for _, child := range node.children {
		child.Remove()
	}
	delete(node.parent.children, node.name)
	node.tree.size--
	return nil
}

func (node *FileNode) String() string {
	return node.name
}

func (node *FileNode) Visit(visiter Visiter) error {
	var keys []string
	for key := range node.children {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, name := range keys {
		child := node.children[name]
		err := child.Visit(visiter)
		if err != nil {
			return err
		}
	}
	return visiter(node)
}


func (node *FileNode) VisitDepthParentFirst(visiter Visiter, evaluator VisitEvaluator) error {
	err := visiter(node)
	if err != nil {
		return err
	}

	var keys []string
	for key := range node.children {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, name := range keys {
		child := node.children[name]
		if evaluator == nil || !evaluator(node) {
			continue
		}
		err = child.VisitDepthParentFirst(visiter, evaluator)
		if err != nil {
			return err
		}
	}
	return err
}

func (node *FileNode) IsWhiteout() bool {
	return strings.HasPrefix(node.name, whiteoutPrefix)
}

func (node *FileNode) Path() string {
	path := []string{}
	curNode := node
	for {
		if curNode.parent == nil {
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


func (node *FileNode) IsLeaf() bool {
	return len(node.children) == 0
}

func (node *FileNode) deriveDiffType(diffType DiffType) error {
	// THE DIFF_TYPE OF A NODE IS ALWAYS THE DIFF_TYPE OF ITS ATTRIBUTES AND ITS CONTENTS
	// THE CONTENTS ARE THE BYTES OF A FILE OR THE CHILDREN OF A DIRECTORY

	if node.IsLeaf() {
		node.AssignDiffType(diffType)
		return nil
	}
	myDiffType := diffType

	for _, v := range node.children {
		vData := v.data
		myDiffType = myDiffType.merge(vData.diffType)

	}
	node.AssignDiffType(myDiffType)
	return nil
}

func (node *FileNode) AssignDiffType(diffType DiffType) error {
	if node.Path() == "/" {
		return nil
	}
	node.data.diffType = diffType
	return nil
}

func (a *FileNode) compare(b *FileNode) DiffType {
	if a == nil && b == nil {
		return Unchanged
	}
	// a is nil but not b
	if a == nil && b != nil {
		return Added
	}

	// b is nil but not a
	if a != nil && b == nil {
		return Removed
	}

	if b.IsWhiteout() {
		return Removed
	}
	if a.name != b.name {
		panic("comparing mismatched nodes")
	}
	// TODO: fails on nil

	return a.data.getDiffType(b.data)
}
