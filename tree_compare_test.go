package main

import "testing"

func TestCompareToTest(t *testing.T) {
	lowerTree := NewTree()
	upperTree := NewTree()
	paths := [...]string{"/etc", "/etc/sudoers", "/etc/hosts", "/usr/bin", "/usr/bin/bash", "/usr"}
	var zeros = [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	for _, value := range paths {
		fakeData := FileChangeInfo{
			path:     value,
			typeflag: 1,
			md5sum:   zeros,
			diffType: nil,
		}

		lowerTree.AddPath(value, fakeData)
		upperTree.AddPath(value, fakeData)
	}
	lowerTree.compareTo(upperTree)
	asserter := func(n *Node) error {
		data, ok := n.data.(FileChangeInfo)
		if !ok {
			t.Errorf("Expecting node with data at %s, but got %+v", n.Path(), n.data)
		}
		if data.diffType == nil {
			t.Errorf("Expected node at %s to have DiffType unchanged, but had nil", n.Path())
		} else if *data.diffType != Unchanged {
			t.Errorf("Expecting node at %s to have DiffType unchanged, but had %v", n.Path(), *data.diffType)
		}
		return nil
	}
	t.Fail()
	lowerTree.Visit(asserter)
}
