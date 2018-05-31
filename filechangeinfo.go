package main

import (
	"fmt"
	"bytes"
)

type FileChangeInfo struct {
	path     string
	typeflag byte
	md5sum   [16]byte
	diffType DiffType
}

type DiffType int

// enum to show whether a file has changed
const (
	Unchanged DiffType = iota
	Changed
	Added
	Removed
)

func (d DiffType) String() string {
	switch d {
	case Unchanged:
		return "Unchanged"
	case Changed:
		return "Changed"
	case Added:
		return "Added"
	case Removed:
		return "Removed"
	default:
		return fmt.Sprintf("%d", int(d))
	}
}

func (a DiffType) merge(b DiffType) DiffType {
	if a == b {
		return a
	}
	return Changed
}

func (a *FileChangeInfo) getDiffType(b *FileChangeInfo) DiffType {
	if a == nil && b == nil {
		return Unchanged
	}
	if a == nil || b == nil {
		return Changed
	}
	if a.typeflag == b.typeflag {
		if bytes.Compare(a.md5sum[:], b.md5sum[:]) == 0 {
			return Unchanged
		}
	}
	return Changed
}


