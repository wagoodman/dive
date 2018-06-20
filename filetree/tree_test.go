package filetree

import (
	"fmt"
	"testing"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func AssertDiffType(node *FileNode, expectedDiffType DiffType) error {
	if node.Data.FileInfo == nil {
		return fmt.Errorf("expected *FileInfo but got nil at Path %s", node.Path())
	}
	if node.Data.DiffType != expectedDiffType {
		return fmt.Errorf("Expecting node at %s to have DiffType %v, but had %v", node.Path(), expectedDiffType, node.Data.DiffType)
	}
	return nil
}

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
	payloadValue := FileInfo{
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

	if *node.Data.FileInfo != payloadValue {
		t.Errorf("Expected '%s' value to be %+v but got %+v", payloadKey, payloadValue, node.Data.FileInfo)
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
		fakeData := FileInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
		lowerTree.AddPath(value, &fakeData)
		upperTree.AddPath(value, &fakeData)
	}
	lowerTree.Compare(upperTree)
	asserter := func(n *FileNode) error {
		if n.Path() == "/" {
			return nil
		}
		if n.Data.FileInfo == nil {
			t.Errorf("Expected *FileInfo but got nil")
			return fmt.Errorf("expected *FileInfo but got nil")
		}
		if (n.Data.DiffType) != Unchanged {
			t.Errorf("Expecting node at %s to have DiffType unchanged, but had %v", n.Path(), n.Data.DiffType)
		}
		return nil
	}
	err := lowerTree.VisitDepthChildFirst(asserter, nil)
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
		lowerTree.AddPath(value, &FileInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		})
	}

	for _, value := range upperPaths {
		upperTree.AddPath(value, &FileInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		})
	}

	failedAssertions := []error{}
	err := lowerTree.Compare(upperTree)
	if err != nil {
		t.Errorf("Expected tree compare to have no errors, got: %v", err)
	}
	asserter := func(n *FileNode) error {

		p := n.Path()
		if p == "/" {
			return nil
		} else if stringInSlice(p,[]string{"/usr/bin/bash"}) {
			if err := AssertDiffType(n, Added); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		} else if stringInSlice(p,[]string{"/usr/bin", "/usr"}) {
			if err := AssertDiffType(n, Changed); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		} else {
			if err := AssertDiffType(n, Unchanged); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		}
		return nil
	}
	err = lowerTree.VisitDepthChildFirst(asserter, nil)
	if err != nil {
		t.Errorf("Expected no errors when visiting nodes, got: %+v", err)
	}

	if len(failedAssertions) > 0 {
		str := "\n"
		for _, value := range failedAssertions {
			str += fmt.Sprintf("  - %s\n", value.Error())
		}
		t.Errorf("Expected no errors when evaluating nodes, got: %s", str)
	}
}

func TestCompareWithChanges(t *testing.T) {
	lowerTree := NewFileTree()
	upperTree := NewFileTree()
	paths := [...]string{"/etc", "/usr", "/etc/hosts", "/etc/sudoers", "/usr/bin"}

	for _, value := range paths {
		lowerTree.AddPath(value, &FileInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		})
		upperTree.AddPath(value, &FileInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0},
		})
	}


	lowerTree.Compare(upperTree)
	failedAssertions := []error{}
	asserter := func(n *FileNode) error {
		p := n.Path()
		if p == "/" {
			return nil
		} else if stringInSlice(p, []string{"/etc", "/usr", "/etc/hosts", "/etc/sudoers", "/usr/bin"}) {
			if err := AssertDiffType(n, Changed); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		} else {
			if err := AssertDiffType(n, Unchanged); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		}
		return nil
	}
	err := lowerTree.VisitDepthChildFirst(asserter, nil)
	if err != nil {
		t.Errorf("Expected no errors when visiting nodes, got: %+v", err)
	}

	if len(failedAssertions) > 0 {
		str := "\n"
		for _, value := range failedAssertions {
			str += fmt.Sprintf("  - %s\n", value.Error())
		}
		t.Errorf("Expected no errors when evaluating nodes, got: %s", str)
	}
}


func TestCompareWithRemoves(t *testing.T) {
	lowerTree := NewFileTree()
	upperTree := NewFileTree()
	lowerPaths := [...]string{"/etc", "/usr", "/etc/hosts", "/etc/sudoers", "/usr/bin", "/root", "/root/example", "/root/example/some1", "/root/example/some2"}
	upperPaths := [...]string{"/.wh.etc", "/usr", "/usr/.wh.bin", "/root/.wh.example"}

	for _, value := range lowerPaths {
		fakeData := FileInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
		lowerTree.AddPath(value, &fakeData)
	}

	for _, value := range upperPaths {
		fakeData := FileInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
		upperTree.AddPath(value, &fakeData)
	}

	lowerTree.Compare(upperTree)
	failedAssertions := []error{}
	asserter := func(n *FileNode) error {
		p := n.Path()
		if p == "/" {
			return nil
		} else if stringInSlice(p,[]string{"/etc", "/usr/bin", "/etc/hosts", "/etc/sudoers", "/root/example/some1", "/root/example/some2", "/root/example"}) {
			if err := AssertDiffType(n, Removed); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		} else if stringInSlice(p,[]string{"/usr", "/root"}) {
			if err := AssertDiffType(n, Changed); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		} else {
			if err := AssertDiffType(n, Unchanged); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		}
		return nil
	}
	err := lowerTree.VisitDepthChildFirst(asserter, nil)
	if err != nil {
		t.Errorf("Expected no errors when visiting nodes, got: %+v", err)
	}

	if len(failedAssertions) > 0 {
		str := "\n"
		for _, value := range failedAssertions {
			str += fmt.Sprintf("  - %s\n", value.Error())
		}
		t.Errorf("Expected no errors when evaluating nodes, got: %s", str)
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
		fakeData := FileInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
		lowerTree.AddPath(value, &fakeData)
	}

	for _, value := range upperPaths {
		fakeData := FileInfo{
			Path:     value,
			Typeflag: 1,
			MD5sum:   [16]byte{1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0},
		}
		upperTree.AddPath(value, &fakeData)
	}
	trees := []*FileTree{lowerTree, upperTree, tree}
	StackRange(trees, 2)
}
