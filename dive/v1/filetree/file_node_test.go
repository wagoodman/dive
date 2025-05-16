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

	one := tree.Root.AddChild("first node!", payload)

	two := tree.Root.AddChild("nil node!", FileInfo{})

	tree.Root.AddChild("third node!", FileInfo{})
	two.AddChild("forth, one level down...", FileInfo{})
	two.AddChild("fifth, one level down...", FileInfo{})
	two.AddChild("fifth, one level down...", FileInfo{})

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

	expectedFC := FileInfo{
		Path: "stufffffs",
	}
	actualFC := one.Data.FileInfo
	if expectedFC.Path != actualFC.Path {
		t.Errorf("Expected 'ones' payload to be %+v got %+v.", expectedFC, actualFC)
	}

}

func TestRemoveChild(t *testing.T) {
	var expected, actual int

	tree := NewFileTree()
	tree.Root.AddChild("first", FileInfo{})
	two := tree.Root.AddChild("nil", FileInfo{})
	tree.Root.AddChild("third", FileInfo{})
	forth := two.AddChild("forth", FileInfo{})
	two.AddChild("fifth", FileInfo{})

	err := forth.Remove()
	checkError(t, err, "unable to setup test")

	expected, actual = 4, tree.Size
	if expected != actual {
		t.Errorf("Expected a tree size of %d got %d.", expected, actual)
	}

	if tree.Root.Children["forth"] != nil {
		t.Errorf("Expected 'forth' node to be deleted.")
	}

	err = two.Remove()
	checkError(t, err, "unable to setup test")

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
	node, _, _ := tree.AddPath(expected, FileInfo{})

	actual := node.Path()
	if expected != actual {
		t.Errorf("Expected path '%s' got '%s'", expected, actual)
	}
}

func TestIsWhiteout(t *testing.T) {
	tree1 := NewFileTree()
	p1, _, _ := tree1.AddPath("/etc/nginx/public1", FileInfo{})
	p2, _, _ := tree1.AddPath("/etc/nginx/.wh.public2", FileInfo{})
	p3, _, _ := tree1.AddPath("/etc/nginx/public3/.wh..wh..opq", FileInfo{})

	if p1.IsWhiteout() != false {
		t.Errorf("Expected path '%s' to **not** be a whiteout file", p1.Name)
	}

	if p2.IsWhiteout() != true {
		t.Errorf("Expected path '%s' to be a whiteout file", p2.Name)
	}

	if p3 != nil {
		t.Errorf("Expected to not be able to add path '%s'", p2.Name)
	}
}

func TestDiffTypeFromAddedChildren(t *testing.T) {
	tree := NewFileTree()
	node, _, _ := tree.AddPath("/usr", *BlankFileChangeInfo("/usr"))
	node.Data.DiffType = Unmodified

	node, _, _ = tree.AddPath("/usr/bin", *BlankFileChangeInfo("/usr/bin"))
	node.Data.DiffType = Added

	node, _, _ = tree.AddPath("/usr/bin2", *BlankFileChangeInfo("/usr/bin2"))
	node.Data.DiffType = Removed

	err := tree.Root.Children["usr"].deriveDiffType(Unmodified)
	checkError(t, err, "unable to setup test")

	if tree.Root.Children["usr"].Data.DiffType != Modified {
		t.Errorf("Expected Modified but got %v", tree.Root.Children["usr"].Data.DiffType)
	}
}
func TestDiffTypeFromRemovedChildren(t *testing.T) {
	tree := NewFileTree()
	_, _, _ = tree.AddPath("/usr", *BlankFileChangeInfo("/usr"))

	info1 := BlankFileChangeInfo("/usr/.wh.bin")
	node, _, _ := tree.AddPath("/usr/.wh.bin", *info1)
	node.Data.DiffType = Removed

	info2 := BlankFileChangeInfo("/usr/.wh.bin2")
	node, _, _ = tree.AddPath("/usr/.wh.bin2", *info2)
	node.Data.DiffType = Removed

	err := tree.Root.Children["usr"].deriveDiffType(Unmodified)
	checkError(t, err, "unable to setup test")

	if tree.Root.Children["usr"].Data.DiffType != Modified {
		t.Errorf("Expected Modified but got %v", tree.Root.Children["usr"].Data.DiffType)
	}

}

func TestDirSize(t *testing.T) {
	tree1 := NewFileTree()
	_, _, err := tree1.AddPath("/etc/nginx/public1", FileInfo{Size: 100})
	checkError(t, err, "unable to setup test")
	_, _, err = tree1.AddPath("/etc/nginx/thing1", FileInfo{Size: 200})
	checkError(t, err, "unable to setup test")
	_, _, err = tree1.AddPath("/etc/nginx/public3/thing2", FileInfo{Size: 300})
	checkError(t, err, "unable to setup test")

	node, _ := tree1.GetNode("/etc/nginx")
	expected, actual := "----------         0:0      600 B ", node.MetadataString()
	if expected != actual {
		t.Errorf("Expected metadata '%s' got '%s'", expected, actual)
	}
}
