package main

import "testing"

func TestAddChild(t *testing.T) {
	var expected, actual int
	tree := NewTree()

	payload := FileChangeInfo{
		path: "stufffffs",
	}

	one := tree.Root().AddChild("first node!", &payload)

	two := tree.Root().AddChild("nil node!", nil)

	tree.Root().AddChild("third node!", nil)
	two.AddChild("forth, one level down...", nil)
	two.AddChild("fifth, one level down...", nil)
	two.AddChild("fifth, one level down...", nil)

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

	expectedFC := &FileChangeInfo{
		path: "stufffffs",
	}
	actualFC := one.data
	if *expectedFC != *actualFC {
		t.Errorf("Expected 'ones' payload to be %+v got %+v.", expectedFC, actualFC)
	}

	if *two.data != *new(FileChangeInfo) {
		t.Errorf("Expected 'twos' payload to be nil got %d.", two.data)
	}



}

func TestRemoveChild(t *testing.T) {
	var expected, actual int

	tree := NewTree()
	tree.Root().AddChild("first", nil)
	two := tree.Root().AddChild("nil", nil)
	tree.Root().AddChild("third", nil)
	forth := two.AddChild("forth", nil)
	two.AddChild("fifth", nil)

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
	two := tree.Root().AddChild("second node!", nil)
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


func TestAddPath(t *testing.T) {
	tree := NewTree()
	tree.AddPath("/etc/nginx/nginx.conf", nil)
	tree.AddPath("/etc/nginx/public", nil)
	tree.AddPath("/var/run/systemd", nil)
	tree.AddPath("/var/run/bashful", nil)
	tree.AddPath("/tmp", nil)
	tree.AddPath("/tmp/nonsense", nil)

	expected := `.
├── etc
│   └── nginx
│       ├── nginx.conf
│       └── public
├── tmp
│   └── nonsense
└── var
    └── run
        ├── bashful
        └── systemd
`
	actual := tree.String()

	if expected != actual {
		t.Errorf("Expected tree string:\n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}

}

func TestRemovePath(t *testing.T) {
	tree := NewTree()
	tree.AddPath("/etc/nginx/nginx.conf", nil)
	tree.AddPath("/etc/nginx/public", nil)
	tree.AddPath("/var/run/systemd", nil)
	tree.AddPath("/var/run/bashful", nil)
	tree.AddPath("/tmp", nil)
	tree.AddPath("/tmp/nonsense", nil)

	tree.RemovePath("/var/run/bashful")
	tree.RemovePath("/tmp")

	expected := `.
├── etc
│   └── nginx
│       ├── nginx.conf
│       └── public
└── var
    └── run
        └── systemd
`
	actual := tree.String()

	if expected != actual {
		t.Errorf("Expected tree string:\n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}

}

func TestPath(t *testing.T) {
	expected := "/etc/nginx/nginx.conf"
	tree := NewTree()
	node, _ := tree.AddPath(expected, nil)

	actual := node.Path()
	if expected != actual {
		t.Errorf("Expected path '%s' got '%s'", expected, actual)
	}
}

func TestIsWhiteout(t *testing.T) {
	tree1 := NewTree()
	p1, _ := tree1.AddPath("/etc/nginx/public1", nil)
	p2, _ := tree1.AddPath("/etc/nginx/.wh.public2", nil)

	if p1.IsWhiteout() != false {
		t.Errorf("Expected path '%s' to **not** be a whiteout file", p1.name)
	}

	if p2.IsWhiteout() != true {
		t.Errorf("Expected path '%s' to be a whiteout file", p2.name)
	}
}


func TestStack(t *testing.T) {
	payloadKey := "/var/run/systemd"
	payloadValue := FileChangeInfo{
		path: "yup",
	}

	tree1 := NewTree()

	tree1.AddPath("/etc/nginx/public", nil)
	tree1.AddPath(payloadKey, nil)
	tree1.AddPath("/var/run/bashful", nil)
	tree1.AddPath("/tmp", nil)
	tree1.AddPath("/tmp/nonsense", nil)

	tree2 := NewTree()
	// add new files
	tree2.AddPath("/etc/nginx/nginx.conf", nil)
	// modify current files
	tree2.AddPath(payloadKey, &payloadValue)
	// whiteout the following files
	tree2.AddPath("/var/run/.wh.bashful", nil)
	tree2.AddPath("/.wh.tmp", nil)

	err := tree1.Stack(tree2)

	if err != nil {
		t.Errorf("Could not stack trees: %v", err)
	}

	expected := `.
├── etc
│   └── nginx
│       ├── nginx.conf
│       └── public
└── var
    └── run
        └── systemd
`

	node, err := tree1.GetNode(payloadKey)
	if err != nil {
		t.Errorf("Expected '%s' to still exist, but it doesn't", payloadKey)
	}

	if *node.data != payloadValue {
		t.Errorf("Expected '%s' value to be %+v but got %+v", payloadKey, payloadValue, node.data)
	}

	actual := tree1.String()

	if expected != actual {
		t.Errorf("Expected tree string:\n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}

}