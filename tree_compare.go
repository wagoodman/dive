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
	// TODO: fails on nil

	return getDiffType(a.data, b.data)
}

func getDiffType(a interface{}, b interface{}) DiffType {
	// they have different types
	if a == nil && b == nil {
		return Unchanged
	}
	if a == nil || b == nil {
		return Changed
	}
	aData, ok := a.(FileChangeInfo)
	if !ok {
		panic(fmt.Errorf("Expected FileChangeInfo but got %+v", aData))
	}
	bData, ok := b.(FileChangeInfo)
	if !ok {
		panic(fmt.Errorf("Expected FileChangeInfo but got %+v", bData))
	}
	if aData.typeflag == bData.typeflag {
		// compare hashes
		if bytes.Compare(aData.md5sum[:], bData.md5sum[:]) == 0 {
			return Unchanged
		}
	}
	return Changed
}

func mergeDiffTypes(a DiffType, b DiffType) DiffType {
	if a == b {
		return a
	}
	return Changed
}

func (tree *FileTree) compareTo(upper *FileTree) error {

	// TODO mark all as unchanged
	markAllUnchanged := func(node *Node) error {
		return node.AssignDiffType(Unchanged)
	}
	err := tree.Visit(markAllUnchanged)
	if err != nil {
		panic(err)
		return err
	}
	graft := func(node *Node) error {
		if node.IsWhiteout() {
			err := tree.MarkRemoved(node.Path())
			if err != nil {
				return fmt.Errorf("Cannot remove node %s: %v", node.Path(), err.Error())
			}
		} else {
			existingNode, _ := tree.GetNode(node.Path())
			if existingNode == nil {
				newNode, err := tree.AddPath(node.Path(), node.data)
				if err != nil {
					return fmt.Errorf("Cannot remove node %s: %v", node.Path(), err.Error())
				}
				newNode.AssignDiffType(Added)
			} else {
				diffType := compareNodes(existingNode, node)
				node.DiffTypeFromChildren(diffType)
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

func (node *Node) IsLeaf() bool {
	return len(node.children) == 0
}

func (node *Node) DiffTypeFromChildren(diffType DiffType) error {
	if node.IsLeaf() {
		node.AssignDiffType(diffType)
		return nil
	}
	myDiffType := diffType

	for _, v := range node.children {
		vData := v.data
		if vData.diffType != nil {
			myDiffType = mergeDiffTypes(myDiffType, *vData.diffType)
		} else {
			return fmt.Errorf("Could not read diffType for node at %s", v.Path())
		}
	}
	node.AssignDiffType(myDiffType)
	return nil
}

func (node *Node) AssignDiffType(diffType DiffType) error {
	if node.Path() == "/" {
		return nil
	}
	node.data.diffType = &diffType
	return nil
}
