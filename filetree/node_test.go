package filetree

import (
	"testing"
)

func TestAddChild(t *testing.T) {
	var expected, actual int
	tree := NewFileTree()

	payload := FileInfo{
		Path: "stufffffs",
	}

	one := tree.Root.AddChild("first node!", &payload)

	two := tree.Root.AddChild("nil node!", nil)

	tree.Root.AddChild("third node!", nil)
	two.AddChild("forth, one level down...", nil)
	two.AddChild("fifth, one level down...", nil)
	two.AddChild("fifth, one level down...", nil)

	expected, actual = 5, tree.Size
	if expected != actual {
		t.Errorf("Expected a tree size of %d got %d.", expected, actual)
	}

	expected, actual = 2, len(two.Children)
	if expected != actual {
		t.Errorf("Expected 'twos' number of children to be %d got %d.", expected, actual)
	}

	expected, actual = 3, len(tree.Root.Children)
	if expected != actual {
		t.Errorf("Expected 'twos' number of children to be %d got %d.", expected, actual)
	}

	expectedFC := &FileInfo{
		Path: "stufffffs",
	}
	actualFC := one.Data.FileInfo
	if *expectedFC != *actualFC {
		t.Errorf("Expected 'ones' payload to be %+v got %+v.", expectedFC, actualFC)
	}

	if two.Data.FileInfo != nil {
		t.Errorf("Expected 'twos' payload to be nil got %+v.", two.Data.FileInfo)
	}

}

func TestRemoveChild(t *testing.T) {
	var expected, actual int

	tree := NewFileTree()
	tree.Root.AddChild("first", nil)
	two := tree.Root.AddChild("nil", nil)
	tree.Root.AddChild("third", nil)
	forth := two.AddChild("forth", nil)
	two.AddChild("fifth", nil)

	forth.Remove()

	expected, actual = 4, tree.Size
	if expected != actual {
		t.Errorf("Expected a tree size of %d got %d.", expected, actual)
	}

	if tree.Root.Children["forth"] != nil {
		t.Errorf("Expected 'forth' node to be deleted.")
	}

	two.Remove()

	expected, actual = 2, tree.Size
	if expected != actual {
		t.Errorf("Expected a tree size of %d got %d.", expected, actual)
	}

	if tree.Root.Children["nil"] != nil {
		t.Errorf("Expected 'nil' node to be deleted.")
	}

}

func TestPath(t *testing.T) {
	expected := "/etc/nginx/nginx.conf"
	tree := NewFileTree()
	node, _ := tree.AddPath(expected, nil)

	actual := node.Path()
	if expected != actual {
		t.Errorf("Expected Path '%s' got '%s'", expected, actual)
	}
}

func TestIsWhiteout(t *testing.T) {
	tree1 := NewFileTree()
	p1, _ := tree1.AddPath("/etc/nginx/public1", nil)
	p2, _ := tree1.AddPath("/etc/nginx/.wh.public2", nil)

	if p1.IsWhiteout() != false {
		t.Errorf("Expected Path '%s' to **not** be a whiteout file", p1.Name)
	}

	if p2.IsWhiteout() != true {
		t.Errorf("Expected Path '%s' to be a whiteout file", p2.Name)
	}
}

func TestDiffTypeFromAddedChildren(t *testing.T) {
	tree := NewFileTree()
	node, _ := tree.AddPath("/usr", BlankFileChangeInfo("/usr"))
	node.Data.DiffType = Unchanged

	info1 := BlankFileChangeInfo("/usr/bin")
	node, _ = tree.AddPath("/usr/bin", info1)
	node.Data.DiffType = Added

	info2 := BlankFileChangeInfo("/usr/bin2")
	node, _ = tree.AddPath("/usr/bin2", info2)
	node.Data.DiffType = Removed

	tree.Root.Children["usr"].deriveDiffType(Unchanged)

	if tree.Root.Children["usr"].Data.DiffType != Changed {
		t.Errorf("Expected Changed but got %v", tree.Root.Children["usr"].Data.DiffType)
	}
}
func TestDiffTypeFromRemovedChildren(t *testing.T) {
	tree := NewFileTree()
	node, _ := tree.AddPath("/usr", BlankFileChangeInfo("/usr"))

	info1 := BlankFileChangeInfo("/usr/.wh.bin")
	node, _ = tree.AddPath("/usr/.wh.bin", info1)
	node.Data.DiffType = Removed

	info2 := BlankFileChangeInfo("/usr/.wh.bin2")
	node, _ = tree.AddPath("/usr/.wh.bin2", info2)
	node.Data.DiffType = Removed

	tree.Root.Children["usr"].deriveDiffType(Unchanged)

	if tree.Root.Children["usr"].Data.DiffType != Changed {
		t.Errorf("Expected Changed but got %v", tree.Root.Children["usr"].Data.DiffType)
	}

}
