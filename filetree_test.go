package main

import (
	"testing"
	"fmt"
)

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
		t.Errorf("Could not stack refTrees: %v", err)
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

func TestCopy(t *testing.T) {
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

	newTree := tree.Copy()
	actual := newTree.String()

	if expected != actual {
		t.Errorf("Expected tree string:\n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}

}



func TestCompareWithNoChanges(t *testing.T) {
	lowerTree := NewTree()
	upperTree := NewTree()
	paths := [...]string{"/etc", "/etc/sudoers", "/etc/hosts", "/usr/bin", "/usr/bin/bash", "/usr"}

	for _, value := range paths {
		fakeData := FileChangeInfo{
			path:     value,
			typeflag: 1,
			md5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			diffType: Unchanged,
		}
		lowerTree.AddPath(value, &fakeData)
		upperTree.AddPath(value, &fakeData)
	}
	lowerTree.compare(upperTree)
	asserter := func(n *FileNode) error {
		if n.Path() == "/" {
			return nil
		}
		if n.data == nil {
			t.Errorf("Expected *FileChangeInfo but got nil")
			return fmt.Errorf("expected *FileChangeInfo but got nil")
		}
		if (n.data.diffType) != Unchanged {
			t.Errorf("Expecting node at %s to have DiffType unchanged, but had %v", n.Path(), n.data.diffType)
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
			diffType: Unchanged,
		}
		lowerTree.AddPath(value, &fakeData)
	}

	for _, value := range upperPaths {
		fakeData := FileChangeInfo{
			path:     value,
			typeflag: 1,
			md5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			diffType: Unchanged,
		}
		upperTree.AddPath(value, &fakeData)
	}

	lowerTree.compare(upperTree)
	asserter := func(n *FileNode) error {

		p := n.Path()
		if p == "/" {
			return nil
		}
		// Adding a file changes the folders it's in
		if p == "/usr/bin/bash" {
			return AssertDiffType(n, Added, t)
		}
		if p == "/usr/bin" {
			return AssertDiffType(n, Changed, t)
		}
		if p == "/usr" {
			return AssertDiffType(n, Changed, t)
		}
		return AssertDiffType(n, Unchanged, t)
	}
	err := lowerTree.Visit(asserter)
	if err != nil {
		t.Error(err)
	}
}

func TestCompareWithChanges(t *testing.T) {
	lowerTree := NewTree()
	upperTree := NewTree()
	lowerPaths := [...]string{"/etc", "/usr", "/etc/hosts", "/etc/sudoers", "/usr/bin"}
	upperPaths := [...]string{"/etc", "/usr", "/etc/hosts", "/etc/sudoers", "/usr/bin"}

	for _, value := range lowerPaths {
		fakeData := FileChangeInfo{
			path:     value,
			typeflag: 1,
			md5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			diffType: Unchanged,
		}
		lowerTree.AddPath(value, &fakeData)
	}

	for _, value := range upperPaths {
		fakeData := FileChangeInfo{
			path:     value,
			typeflag: 1,
			md5sum:   [16]byte{1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0},
			diffType: Unchanged,
		}
		upperTree.AddPath(value, &fakeData)
	}

	lowerTree.compare(upperTree)
	asserter := func(n *FileNode) error {
		p := n.Path()
		if p == "/" {
			return nil
		}
		return AssertDiffType(n, Changed, t)
	}
	err := lowerTree.Visit(asserter)
	if err != nil {
		t.Error(err)
	}
}

