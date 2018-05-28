package main

import "testing"

func compareToTest(t *testing.T) {
	lowerTree := NewTree()
	upperTree := NewTree()
	paths := [5]string{"/etc", "/etc/sudoers", "/etc/hosts", "/usr/bin", "/usr/bin/bash"}
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
		data := n.data.(FileChangeInfo)
		if data.diffType == nil {
			t.Errorf("Expected node at %s to have DiffType unchanged, but had nil", n.Path())
		}
		if *data.diffType != Unchanged {
			t.Errorf("Expecting node at %s to have DiffType unchanged, but had %v", n.Path(), *data.diffType)
		}
		return nil
	}
	lowerTree.Visit(asserter)
}
