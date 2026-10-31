package filetree

import (
	"testing"

	"github.com/fatih/color"
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

	if p1.IsWhiteout() != false {
		t.Errorf("Expected path '%s' to **not** be a whiteout file", p1.Name)
	}

	if p2.IsWhiteout() != true {
		t.Errorf("Expected path '%s' to be a whiteout file", p2.Name)
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

func checkError(t *testing.T, err error, message string) {
	if err != nil {
		t.Errorf(message+": %+v", err)
	}
}

func TestDiffTypeColors(t *testing.T) {
	defer ResetDiffTypeColors()

	// Verify defaults
	if DiffTypeColor(Added) == nil {
		t.Errorf("Expected default color for Added to be non-nil")
	}
	if DiffTypeColor(Removed) == nil {
		t.Errorf("Expected default color for Removed to be non-nil")
	}
	if DiffTypeColor(Modified) == nil {
		t.Errorf("Expected default color for Modified to be non-nil")
	}
	if DiffTypeColor(Unmodified) == nil {
		t.Errorf("Expected default color for Unmodified to be non-nil")
	}

	// Test custom color override
	custom := color.New(color.FgCyan)
	SetDiffTypeColor(Added, custom)
	if DiffTypeColor(Added) != custom {
		t.Errorf("Expected custom color for Added")
	}

	// Test nil does not override
	SetDiffTypeColor(Added, nil)
	if DiffTypeColor(Added) != custom {
		t.Errorf("Expected custom color to remain after passing nil")
	}

	// Test reset
	ResetDiffTypeColors()
	if DiffTypeColor(Added) == custom {
		t.Errorf("Expected color to be reset back to default")
	}
}
