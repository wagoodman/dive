package filetree

import (
	"fmt"
	"testing"
)

func TestPrintTree(t *testing.T) {
	tree := NewFileTree()
	tree.Root.AddChild("first node!", nil)
	two := tree.Root.AddChild("second node!", nil)
	tree.Root.AddChild("third node!", nil)
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
	tree := NewFileTree()
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
	tree := NewFileTree()
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
		Path: "yup",
	}

	tree1 := NewFileTree()

	tree1.AddPath("/etc/nginx/public", nil)
	tree1.AddPath(payloadKey, nil)
	tree1.AddPath("/var/run/bashful", nil)
	tree1.AddPath("/tmp", nil)
	tree1.AddPath("/tmp/nonsense", nil)

	tree2 := NewFileTree()
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

	if *node.Data != payloadValue {
		t.Errorf("Expected '%s' value to be %+v but got %+v", payloadKey, payloadValue, node.Data)
	}

	actual := tree1.String()

	if expected != actual {
		t.Errorf("Expected tree string:\n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}

}

func TestCopy(t *testing.T) {
	tree := NewFileTree()
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

	NewFileTree := tree.Copy()
	actual := NewFileTree.String()

	if expected != actual {
		t.Errorf("Expected tree string:\n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}

}

func TestCompareWithNoChanges(t *testing.T) {
	lowerTree := NewFileTree()
	upperTree := NewFileTree()
	paths := [...]string{"/etc", "/etc/sudoers", "/etc/hosts", "/usr/bin", "/usr/bin/bash", "/usr"}

	for _, value := range paths {
		fakeData := FileChangeInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			DiffType: Unchanged,
		}
		lowerTree.AddPath(value, &fakeData)
		upperTree.AddPath(value, &fakeData)
	}
	lowerTree.Compare(upperTree)
	asserter := func(n *FileNode) error {
		if n.Path() == "/" {
			return nil
		}
		if n.Data == nil {
			t.Errorf("Expected *FileChangeInfo but got nil")
			return fmt.Errorf("expected *FileChangeInfo but got nil")
		}
		if (n.Data.DiffType) != Unchanged {
			t.Errorf("Expecting node at %s to have DiffType unchanged, but had %v", n.Path(), n.Data.DiffType)
		}
		return nil
	}
	err := lowerTree.Visit(asserter)
	if err != nil {
		t.Error(err)
	}
}

func TestCompareWithAdds(t *testing.T) {
	lowerTree := NewFileTree()
	upperTree := NewFileTree()
	lowerPaths := [...]string{"/etc", "/etc/sudoers", "/usr", "/etc/hosts", "/usr/bin"}
	upperPaths := [...]string{"/etc", "/etc/sudoers", "/usr", "/etc/hosts", "/usr/bin", "/usr/bin/bash"}

	for _, value := range lowerPaths {
		fakeData := FileChangeInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			DiffType: Unchanged,
		}
		lowerTree.AddPath(value, &fakeData)
	}

	for _, value := range upperPaths {
		fakeData := FileChangeInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			DiffType: Unchanged,
		}
		upperTree.AddPath(value, &fakeData)
	}

	lowerTree.Compare(upperTree)
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
	lowerTree := NewFileTree()
	upperTree := NewFileTree()
	lowerPaths := [...]string{"/etc", "/usr", "/etc/hosts", "/etc/sudoers", "/usr/bin"}
	upperPaths := [...]string{"/etc", "/usr", "/etc/hosts", "/etc/sudoers", "/usr/bin"}

	for _, value := range lowerPaths {
		fakeData := FileChangeInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			DiffType: Unchanged,
		}
		lowerTree.AddPath(value, &fakeData)
	}

	for _, value := range upperPaths {
		fakeData := FileChangeInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0},
			DiffType: Unchanged,
		}
		upperTree.AddPath(value, &fakeData)
	}

	lowerTree.Compare(upperTree)
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

func TestStackRange(t *testing.T) {
	tree := NewFileTree()
	tree.AddPath("/etc/nginx/nginx.conf", nil)
	tree.AddPath("/etc/nginx/public", nil)
	tree.AddPath("/var/run/systemd", nil)
	tree.AddPath("/var/run/bashful", nil)
	tree.AddPath("/tmp", nil)
	tree.AddPath("/tmp/nonsense", nil)

	tree.RemovePath("/var/run/bashful")
	tree.RemovePath("/tmp")

	lowerTree := NewFileTree()
	upperTree := NewFileTree()
	lowerPaths := [...]string{"/etc", "/usr", "/etc/hosts", "/etc/sudoers", "/usr/bin"}
	upperPaths := [...]string{"/etc", "/usr", "/etc/hosts", "/etc/sudoers", "/usr/bin"}

	for _, value := range lowerPaths {
		fakeData := FileChangeInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			DiffType: Unchanged,
		}
		lowerTree.AddPath(value, &fakeData)
	}

	for _, value := range upperPaths {
		fakeData := FileChangeInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0},
			DiffType: Unchanged,
		}
		upperTree.AddPath(value, &fakeData)
	}
	trees := []*FileTree{lowerTree, upperTree, tree}
	StackRange(trees, 2)
}
