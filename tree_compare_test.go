package main

import (
	"fmt"
	"testing"
)

func TestCompareWithNoChanges(t *testing.T) {
	lowerTree := NewTree()
	upperTree := NewTree()
	paths := [...]string{"/etc", "/etc/sudoers", "/etc/hosts", "/usr/bin", "/usr/bin/bash", "/usr"}

	for _, value := range paths {
		fakeData := FileChangeInfo{
			path:     value,
			typeflag: 1,
			md5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			diffType: Unchanged,
		}
		lowerTree.AddPath(value, &fakeData)
		upperTree.AddPath(value, &fakeData)
	}
	lowerTree.compareTo(upperTree)
	asserter := func(n *Node) error {
		if n.Path() == "/" {
			return nil
		}
		if n.data == nil {
			t.Errorf("Expected *FileChangeInfo but got nil")
			return fmt.Errorf("expected *FileChangeInfo but got nil")
		}
		if (n.data.diffType) != Unchanged {
			t.Errorf("Expecting node at %s to have DiffType unchanged, but had %v", n.Path(), n.data.diffType)
		}
		return nil
	}
	err := lowerTree.Visit(asserter)
	if err != nil {
		t.Error(err)
	}
}

func TestCompareWithAdds(t *testing.T) {
	lowerTree := NewTree()
	upperTree := NewTree()
	lowerPaths := [...]string{"/etc", "/etc/sudoers", "/usr", "/etc/hosts", "/usr/bin"}
	upperPaths := [...]string{"/etc", "/etc/sudoers", "/usr", "/etc/hosts", "/usr/bin", "/usr/bin/bash"}

	for _, value := range lowerPaths {
		fakeData := FileChangeInfo{
			path:     value,
			typeflag: 1,
			md5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			diffType: Unchanged,
		}
		lowerTree.AddPath(value, &fakeData)
	}

	for _, value := range upperPaths {
		fakeData := FileChangeInfo{
			path:     value,
			typeflag: 1,
			md5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			diffType: Unchanged,
		}
		upperTree.AddPath(value, &fakeData)
	}

	lowerTree.compareTo(upperTree)
	asserter := func(n *Node) error {

		p := n.Path()
		if p == "/" {
			return nil
		}
		// Adding a file changes the folders it's in
		if p == "/usr/bin/bash" {
			return AssertDiffType(n, Added, t)
		}
		if p == "/usr/bin" {
			return AssertDiffType(n, Changed, t)
		}
		if p == "/usr" {
			return AssertDiffType(n, Changed, t)
		}
		return AssertDiffType(n, Unchanged, t)
	}
	err := lowerTree.Visit(asserter)
	if err != nil {
		t.Error(err)
	}
}

func TestCompareWithChanges(t *testing.T) {
	lowerTree := NewTree()
	upperTree := NewTree()
	lowerPaths := [...]string{"/etc", "/usr", "/etc/hosts", "/etc/sudoers", "/usr/bin"}
	upperPaths := [...]string{"/etc", "/usr", "/etc/hosts", "/etc/sudoers", "/usr/bin"}

	for _, value := range lowerPaths {
		fakeData := FileChangeInfo{
			path:     value,
			typeflag: 1,
			md5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			diffType: Unchanged,
		}
		lowerTree.AddPath(value, &fakeData)
	}

	for _, value := range upperPaths {
		fakeData := FileChangeInfo{
			path:     value,
			typeflag: 1,
			md5sum:   [16]byte{1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0},
			diffType: Unchanged,
		}
		upperTree.AddPath(value, &fakeData)
	}

	lowerTree.compareTo(upperTree)
	asserter := func(n *Node) error {
		p := n.Path()
		if p == "/" {
			return nil
		}
		return AssertDiffType(n, Changed, t)
	}
	err := lowerTree.Visit(asserter)
	if err != nil {
		t.Error(err)
	}
}

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
	merged := mergeDiffTypes(a, b)
	if merged != Unchanged {
		t.Errorf("Expected Unchaged (0) but got %v", merged)
	}
	a = Changed
	b = Unchanged
	merged = mergeDiffTypes(a, b)
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
	tree.root.children["usr"].DiffTypeFromChildren(Unchanged)
	if tree.root.children["usr"].data.diffType != Changed {
		t.Errorf("Expected Changed but got %v", tree.root.children["usr"].data.diffType)
	}
}

func AssertDiffType(node *Node, expectedDiffType DiffType, t *testing.T) error {
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
