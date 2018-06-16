package filetree

import (
	"fmt"
	"testing"
)

func TestAssignDiffType(t *testing.T) {
	tree := NewFileTree()
	tree.AddPath("/usr", BlankFileChangeInfo("/usr", Changed))
	if tree.Root.Children["usr"].Data.DiffType != Changed {
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

func AssertDiffType(node *FileNode, expectedDiffType DiffType, t *testing.T) error {
	if node.Data.FileInfo == nil {
		t.Errorf("Expected *FileInfo but got nil at Path %s", node.Path())
		return fmt.Errorf("expected *FileInfo but got nil at Path %s", node.Path())
	}
	if node.Data.DiffType != expectedDiffType {
		t.Errorf("Expecting node at %s to have DiffType %v, but had %v", node.Path(), expectedDiffType, node.Data.DiffType)
		return fmt.Errorf("Assertion failed")
	}
	return nil
}

func BlankFileChangeInfo(path string, diffType DiffType) (f *FileInfo) {
	result := FileInfo{
		Path:     path,
		Typeflag: 1,
		MD5sum:   [16]byte{1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0},
	}
	return &result
}
