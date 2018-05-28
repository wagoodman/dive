package main

import (
	"bytes"
	"fmt"
)

func differ() func(int) int {
	sum := 0
	return func(x int) int {
		sum += x
		return sum
	}
}

type DiffType int

// enum to show whether a file has changed
const (
	Unchanged DiffType = iota
	Changed
	Added
	Removed
)

func compareNodes(a *Node, b *Node) DiffType {
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
	return getDiffType(a.data.(FileChangeInfo), b.data.(FileChangeInfo))
}

func getDiffType(a FileChangeInfo, b FileChangeInfo) DiffType {
	// they have different types
	if a.typeflag == b.typeflag {
		// compare hashes
		if bytes.Compare(a.md5sum[:], b.md5sum[:]) == 0 {
			return Unchanged
		}
	}
	return Changed
}

func (tree *FileTree) compareTo(upper *FileTree) error {

	// TODO mark all as unchanged
	graft := func(node *Node) error {
		if node.IsWhiteout() {
			err := tree.MarkRemoved(node.Path())
			if err != nil {
				return fmt.Errorf("Cannot remove node %s: %v", node.Path(), err.Error())
			}
		} else {
			existingNode, _ := tree.GetNode(node.Path())
			diffType := compareNodes(existingNode, node)
			if node.IsLeaf() {
				node.AssignDiffType(diffType)
			} else {
				node.DiffTypeFromChildren()
			}
			// TODO mark diff type
			newNode, err := tree.AddPath(node.Path(), node.data)
			if err != nil {
				return fmt.Errorf("Cannot add node %s: %v", newNode.Path(), err.Error())
			}
		}
		return nil
	}
	return upper.Visit(graft)
}

// THE DIFF_TYPE OF A NODE IS ALWAYS THE DIFF_TYPE OF ITS ATTRIBUTES AND ITS CONTENTS
// THE CONTENTS ARE THE BYTES OF A FILE OR THE CHILDREN OF A DIRECTORY

func (tree *FileTree) MarkRemoved(path string) error {
	node, err := tree.GetNode(path)
	if err != nil {
		return err
	}
	return node.AssignDiffType(Removed)
}

func (node *Node) IsEmpty() bool {
	return len(node.children) == 0
}

func (node *Node) DiffTypeFromChildren() {

	for i := 2; i < n; i++ {
		ins[0] = out
		ins[1] = in.Index(i)
		out = fn.Call(ins[:])[0]
	}
}

func (node *Node) AssignDiffType(diffType DiffType) error {
	f, ok := node.data.(FileChangeInfo)
	if ok {
		f.diffType = &diffType
	}
	return fmt.Errorf("Cannot assign diffType on %v because a type assertion failed", node.data)
}

type DiffTree struct {
	root *Node
	size int
	name string
}
