package filetree

import (
	"testing"
)

func TestAssignDiffType(t *testing.T) {
	tree := NewFileTree()
	node, _, err := tree.AddPath("/usr", *BlankFileChangeInfo("/usr"))
	if err != nil {
		t.Errorf("Expected no error from fetching path. got: %v", err)
	}
	node.Data.DiffType = Changed
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

func BlankFileChangeInfo(path string) (f *FileInfo) {
	result := FileInfo{
		Path:     path,
		TypeFlag: 1,
		hash:     123,
	}
	return &result
}
