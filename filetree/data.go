package filetree

import (
	"archive/tar"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
)

// enum to show whether a file has changed
const (
	Unchanged DiffType = iota
	Changed
	Added
	Removed
)

type NodeData struct {
	ViewInfo  ViewInfo
	FileInfo  *FileInfo
	DiffType  DiffType
}

type ViewInfo struct {
	Collapsed bool
	Hidden    bool
}

type FileInfo struct {
	Path     string
	Typeflag byte
	MD5sum   [16]byte
}

type DiffType int

func NewNodeData() (*NodeData) {
	return &NodeData{
		ViewInfo: *NewViewInfo(),
		FileInfo: nil,
		DiffType: Unchanged,
	}
}

func (data *NodeData) Copy() (*NodeData) {
	return &NodeData{
		ViewInfo: *data.ViewInfo.Copy(),
		FileInfo: data.FileInfo.Copy(),
		DiffType: data.DiffType,
	}
}


func NewViewInfo() (view *ViewInfo) {
	return &ViewInfo{
		Collapsed: false,
		Hidden: false,
	}
}

func (view *ViewInfo) Copy() (newView *ViewInfo) {
	newView = NewViewInfo()
	*newView = *view
	return newView
}

func NewFileInfo(reader *tar.Reader, header *tar.Header, path string) FileInfo {
	if header.Typeflag == tar.TypeDir {
		return FileInfo{
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
	return FileInfo{
		Path:     path,
		Typeflag: header.Typeflag,
		MD5sum:   md5.Sum(fileBytes),
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

func (data *FileInfo) Copy() *FileInfo {
	if data == nil {
		return nil
	}
	return &FileInfo{
		Path:     data.Path,
		Typeflag: data.Typeflag,
		MD5sum:   data.MD5sum,
	}
}

func (data *FileInfo) getDiffType(other *FileInfo) DiffType {
	if data == nil && other == nil {
		return Unchanged
	}
	if data == nil || other == nil {
		return Changed
	}
	if data.Typeflag == other.Typeflag {
		if bytes.Compare(data.MD5sum[:], other.MD5sum[:]) == 0 {
			return Unchanged
		}
	}
	return Changed
}
