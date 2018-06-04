package main

import (
	"fmt"
	"testing"
)

func TestAssignDiffType(t *testing.T) {
	tree := NewTree()
	tree.AddPath("/usr", BlankFileChangeInfo("/usr", Changed))
	if tree.root.children["usr"].data.diffType != Changed {
		t.Fail()
	}
}

func TestMergeDiffTypes(t *testing.T) {
	a := Unchanged
	b := Unchanged
	merged := a.merge(b)
	if merged != Unchanged {
		t.Errorf("Expected Unchaged (0) but got %v", merged)
	}
	a = Changed
	b = Unchanged
	merged = a.merge(b)
	if merged != Changed {
		t.Errorf("Expected Unchaged (0) but got %v", merged)
	}
}

func TestDiffTypeFromChildren(t *testing.T) {
	tree := NewTree()
	tree.AddPath("/usr", BlankFileChangeInfo("/usr", Unchanged))
	info1 := BlankFileChangeInfo("/usr/bin", Added)
	tree.AddPath("/usr/bin", info1)
	info2 := BlankFileChangeInfo("/usr/bin2", Removed)
	tree.AddPath("/usr/bin2", info2)
	tree.root.children["usr"].deriveDiffType(Unchanged)
	if tree.root.children["usr"].data.diffType != Changed {
		t.Errorf("Expected Changed but got %v", tree.root.children["usr"].data.diffType)
	}
}

func AssertDiffType(node *FileNode, expectedDiffType DiffType, t *testing.T) error {
	if node.data == nil {
		t.Errorf("Expected *FileChangeInfo but got nil at path %s", node.Path())
		return fmt.Errorf("expected *FileChangeInfo but got nil at path %s", node.Path())
	}
	if node.data.diffType != expectedDiffType {
		t.Errorf("Expecting node at %s to have DiffType %v, but had %v", node.Path(), expectedDiffType, node.data.diffType)
		return fmt.Errorf("Assertion failed")
	}
	return nil
}

func BlankFileChangeInfo(path string, diffType DiffType) (f *FileChangeInfo) {
	result := FileChangeInfo{
		path:     path,
		typeflag: 1,
		md5sum:   [16]byte{1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0},
		diffType: diffType,
	}
	return &result
}
