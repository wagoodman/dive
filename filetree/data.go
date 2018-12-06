package filetree

import (
	"archive/tar"
	"fmt"
	"io"

	"github.com/cespare/xxhash"
	"github.com/sirupsen/logrus"
)

const (
	Unchanged DiffType = iota
	Changed
	Added
	Removed
)

var GlobalFileTreeCollapse bool

// NewNodeData creates an empty NodeData struct for a FileNode
func NewNodeData() *NodeData {
	return &NodeData{
		ViewInfo: *NewViewInfo(),
		FileInfo: FileInfo{},
		DiffType: Unchanged,
	}
}

// Copy duplicates a NodeData
func (data *NodeData) Copy() *NodeData {
	return &NodeData{
		ViewInfo: *data.ViewInfo.Copy(),
		FileInfo: *data.FileInfo.Copy(),
		DiffType: data.DiffType,
	}
}

// NewViewInfo creates a default ViewInfo
func NewViewInfo() (view *ViewInfo) {
	return &ViewInfo{
		Collapsed: GlobalFileTreeCollapse,
		Hidden:    false,
	}
}

// Copy duplicates a ViewInfo
func (view *ViewInfo) Copy() (newView *ViewInfo) {
	newView = NewViewInfo()
	*newView = *view
	return newView
}

func getHashFromReader(reader io.Reader) uint64 {
	h := xxhash.New()

	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			logrus.Panic(err)
		}
		if n == 0 {
			break
		}

		h.Write(buf[:n])
	}

	return h.Sum64()
}

// NewFileInfo extracts the metadata from a tar header and file contents and generates a new FileInfo object.
func NewFileInfo(reader *tar.Reader, header *tar.Header, path string) FileInfo {
	if header.Typeflag == tar.TypeDir {
		return FileInfo{
			Path:     path,
			TypeFlag: header.Typeflag,
			Linkname: header.Linkname,
			hash:     0,
			Size:     header.FileInfo().Size(),
			Mode:     header.FileInfo().Mode(),
			Uid:      header.Uid,
			Gid:      header.Gid,
			IsDir:    header.FileInfo().IsDir(),
		}
	}

	hash := getHashFromReader(reader)

	return FileInfo{
		Path:     path,
		TypeFlag: header.Typeflag,
		Linkname: header.Linkname,
		hash:     hash,
		Size:     header.FileInfo().Size(),
		Mode:     header.FileInfo().Mode(),
		Uid:      header.Uid,
		Gid:      header.Gid,
		IsDir:    header.FileInfo().IsDir(),
	}
}

// Copy duplicates a FileInfo
func (data *FileInfo) Copy() *FileInfo {
	if data == nil {
		return nil
	}
	return &FileInfo{
		Path:     data.Path,
		TypeFlag: data.TypeFlag,
		Linkname: data.Linkname,
		hash:     data.hash,
		Size:     data.Size,
		Mode:     data.Mode,
		Uid:      data.Uid,
		Gid:      data.Gid,
		IsDir:    data.IsDir,
	}
}

// Compare determines the DiffType between two FileInfos based on the type and contents of each given FileInfo
func (data *FileInfo) Compare(other FileInfo) DiffType {
	if data.TypeFlag == other.TypeFlag {
		if data.hash == other.hash {
			return Unchanged
		}
	}
	return Changed
}

// String of a DiffType
func (diff DiffType) String() string {
	switch diff {
	case Unchanged:
		return "Unchanged"
	case Changed:
		return "Changed"
	case Added:
		return "Added"
	case Removed:
		return "Removed"
	default:
		return fmt.Sprintf("%d", int(diff))
	}
}

// merge two DiffTypes into a single result. Essentially, return the given value unless they two values differ,
// in which case we can only determine that there is "a change".
func (diff DiffType) merge(other DiffType) DiffType {
	if diff == other {
		return diff
	}
	return Changed
}
