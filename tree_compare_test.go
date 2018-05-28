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
			diffType: nil,
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
		if n.data.diffType == nil {
			t.Errorf("Expected node at %s to have DiffType unchanged, but had nil", n.Path())
		} else if *(n.data.diffType) != Unchanged {
			t.Errorf("Expecting node at %s to have DiffType unchanged, but had %v", n.Path(), *n.data.diffType)
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
			diffType: nil,
		}
		lowerTree.AddPath(value, &fakeData)
	}

	for _, value := range upperPaths {
		fakeData := FileChangeInfo{
			path:     value,
			typeflag: 1,
			md5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			diffType: nil,
		}
		upperTree.AddPath(value, &fakeData)
	}

	lowerTree.compareTo(upperTree)
	asserter := func(n *Node) error {

		p := n.Path()
		if p == "/" {
			return nil
		}
		if p == "/usr/bin" || p == "/usr/bin/bash" {
			return AssertDiffType(n, Added, t)
		} else {
			return AssertDiffType(n, Unchanged, t)
		}
	}
	err := lowerTree.Visit(asserter)
	if err != nil {
		t.Error(err)
	}
}

func AssertDiffType(node *Node, expectedDiffType DiffType, t *testing.T) error {
	if node.data == nil {
		t.Errorf("Expected *FileChangeInfo but got nil at path %s", node.Path())
		return fmt.Errorf("expected *FileChangeInfo but got nil at path %s", node.Path())
	}
	if node.data.diffType == nil {
		t.Errorf("Expected node at %s to have DiffType Added, but had nil", node.Path())
		return fmt.Errorf("Assertion failed")
	} else if *(node.data.diffType) != expectedDiffType {
		t.Errorf("Expecting node at %s to have DiffType Added, but had %v", node.Path(), *node.data.diffType)
		return fmt.Errorf("Assertion failed")
	}
	return nil
}
