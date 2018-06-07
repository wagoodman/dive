package filetree

import (
	"archive/tar"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
)

type FileChangeInfo struct {
	Path     string
	Typeflag byte
	MD5sum   [16]byte
	DiffType DiffType
}

type DiffType int

// enum to show whether a file has changed
const (
	Unchanged DiffType = iota
	Changed
	Added
	Removed
)

func NewFileChangeInfo(reader *tar.Reader, header *tar.Header, path string) FileChangeInfo {
	if header.Typeflag == tar.TypeDir {
		return FileChangeInfo{
			Path:     path,
			Typeflag: header.Typeflag,
			MD5sum:   [16]byte{},
		}
	}
	fileBytes := make([]byte, header.Size)
	_, err := reader.Read(fileBytes)
	if err != nil && err != io.EOF {
		panic(err)
	}
	return FileChangeInfo{
		Path:     path,
		Typeflag: header.Typeflag,
		MD5sum:   md5.Sum(fileBytes),
		DiffType: Unchanged,
	}
}

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
	if a.Typeflag == b.Typeflag {
		if bytes.Compare(a.MD5sum[:], b.MD5sum[:]) == 0 {
			return Unchanged
		}
	}
	return Changed
}
