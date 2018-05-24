package main

import "testing"

func TestAddChild(t *testing.T) {
	var expected, actual int
	tree := NewTree()

	_, err := tree.Root().AddChild("first node!", 1)
	if err != nil {
		t.Errorf("Adding valued child should not result in error.")
	}

	two, err := tree.Root().AddChild("nil node!", nil)
	if err != nil {
		t.Errorf("Adding nil child should not result in error.")
	}

	tree.Root().AddChild("third node!", 3)
	two.AddChild("forth, one level down...", 4)
	two.AddChild("fifth, one level down...", 5)

	_, err = two.AddChild("fifth, one level down...", 5)
	if err == nil {
		t.Errorf("Expected an error when adding duplicate nodes, no error given.")
	}

	expected, actual = 5, tree.size
	if expected != actual {
		t.Errorf("Expected a tree size of %d got %d.", expected, actual)
	}

	expected, actual = 2, len(two.children)
	if expected != actual {
		t.Errorf("Expected 'twos' number of children to be %d got %d.", expected, actual)
	}

	expected, actual = 3, len(tree.Root().children)
	if expected != actual {
		t.Errorf("Expected 'twos' number of children to be %d got %d.", expected, actual)
	}

}

func TestRemoveChild(t *testing.T) {
	var expected, actual int

	tree := NewTree()
	tree.Root().AddChild("first", 1)
	two, _ := tree.Root().AddChild("nil", nil)
	tree.Root().AddChild("third", 3)
	forth, _ := two.AddChild("forth", 4)
	two.AddChild("fifth", 5)

	forth.Remove()

	expected, actual = 4, tree.size
	if expected != actual {
		t.Errorf("Expected a tree size of %d got %d.", expected, actual)
	}

	if tree.Root().children["forth"] != nil {
		t.Errorf("Expected 'forth' node to be deleted.")
	}

	two.Remove()

	expected, actual = 2, tree.size
	if expected != actual {
		t.Errorf("Expected a tree size of %d got %d.", expected, actual)
	}

	if tree.Root().children["nil"] != nil {
		t.Errorf("Expected 'nil' node to be deleted.")
	}

}

func TestPrintTree(t *testing.T) {
	tree := NewTree()
	tree.Root().AddChild("first node!", nil)
	two, _ := tree.Root().AddChild("second node!", nil)
	tree.Root().AddChild("third node!", nil)
	two.AddChild("forth, one level down...", nil)

	expected := `.
├── first node!
├── second node!
│   └── forth, one level down...
└── third node!
`
	actual := tree.String()

	if expected != actual {
		t.Errorf("Expected tree string:\n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}

}
