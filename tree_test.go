package main

import "testing"

func TestAddChild(t *testing.T) {
	var expected, actual int
	tree := NewTree()

	one := tree.Root().AddChild("first node!", 1)

	two := tree.Root().AddChild("nil node!", nil)

	tree.Root().AddChild("third node!", 3)
	two.AddChild("forth, one level down...", 4)
	two.AddChild("fifth, one level down...", 5)
	two.AddChild("fifth, one level down...", 5)

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

	if two.data != nil {
		t.Errorf("Expected 'ones' payload to be nil got %d.", two.data)
	}

	expected, actual = 1, one.data.(int)
	if expected != actual {
		t.Errorf("Expected 'ones' payload to be %d got %d.", expected, actual)
	}

}

func TestRemoveChild(t *testing.T) {
	var expected, actual int

	tree := NewTree()
	tree.Root().AddChild("first", 1)
	two := tree.Root().AddChild("nil", nil)
	tree.Root().AddChild("third", 3)
	forth := two.AddChild("forth", 4)
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
	tree.AddPath("/etc/nginx/nginx.conf", 1)
	tree.AddPath("/etc/nginx/public", 2)
	tree.AddPath("/var/run/systemd", 3)
	tree.AddPath("/var/run/bashful", 4)
	tree.AddPath("/tmp", 5)
	tree.AddPath("/tmp/nonsense", 6)

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
	tree.AddPath("/etc/nginx/nginx.conf", 1)
	tree.AddPath("/etc/nginx/public", 2)
	tree.AddPath("/var/run/systemd", 3)
	tree.AddPath("/var/run/bashful", 4)
	tree.AddPath("/tmp", 5)
	tree.AddPath("/tmp/nonsense", 6)

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
	p1, _ := tree1.AddPath("/etc/nginx/public1", 2)
	p2, _ := tree1.AddPath("/etc/nginx/.wh.public2", 2)

	if p1.IsWhiteout() != false {
		t.Errorf("Expected path '%s' to **not** be a whiteout file", p1.name)
	}

	if p2.IsWhiteout() != true {
		t.Errorf("Expected path '%s' to be a whiteout file", p2.name)
	}
}


func TestStack(t *testing.T) {
	payloadKey := "/var/run/systemd"
	payloadValue := 1263487

	tree1 := NewTree()

	tree1.AddPath("/etc/nginx/public", 2)
	tree1.AddPath(payloadKey, 3)
	tree1.AddPath("/var/run/bashful", 4)
	tree1.AddPath("/tmp", 5)
	tree1.AddPath("/tmp/nonsense", 6)

	tree2 := NewTree()
	// add new files
	tree2.AddPath("/etc/nginx/nginx.conf", 1)
	// modify current files
	tree2.AddPath(payloadKey, payloadValue)
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

	if node.data != payloadValue {
		t.Errorf("Expected '%s' value to be %d but got %d", payloadKey, payloadValue, node.data.(int))
	}

	actual := tree1.String()

	if expected != actual {
		t.Errorf("Expected tree string:\n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}

}