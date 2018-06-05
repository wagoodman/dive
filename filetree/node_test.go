package filetree

import "testing"

func TestAddChild(t *testing.T) {
	var expected, actual int
	tree := NewFileTree()

	payload := FileChangeInfo{
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

	expectedFC := &FileChangeInfo{
		Path: "stufffffs",
	}
	actualFC := one.Data
	if *expectedFC != *actualFC {
		t.Errorf("Expected 'ones' payload to be %+v got %+v.", expectedFC, actualFC)
	}

	if *two.Data != *new(FileChangeInfo) {
		t.Errorf("Expected 'twos' payload to be nil got %+v.", two.Data)
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
