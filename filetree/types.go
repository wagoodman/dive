package filetree

import (
	"github.com/google/uuid"
	"os"
)

// FileTree represents a set of files, directories, and their relations.
type FileTree struct {
	Root     *FileNode
	Size     int
	FileSize uint64
	Name     string
	Id       uuid.UUID
}

// FileNode represents a single file, its relation to files beneath it, the tree it exists in, and the metadata of the given file.
type FileNode struct {
	Tree     *FileTree
	Parent   *FileNode
	Name     string
	Data     NodeData
	Children map[string]*FileNode
	path     string
}

// NodeData is the payload for a FileNode
type NodeData struct {
	ViewInfo ViewInfo
	FileInfo FileInfo
	DiffType DiffType
}

// ViewInfo contains UI specific detail for a specific FileNode
type ViewInfo struct {
	Collapsed bool
	Hidden    bool
}

// FileInfo contains tar metadata for a specific FileNode
type FileInfo struct {
	Path     string
	TypeFlag byte
	Linkname string
	hash     uint64
	Size     int64
	Mode     os.FileMode
	Uid      int
	Gid      int
	IsDir    bool
}

// DiffType defines the comparison result between two FileNodes
type DiffType int

// EfficiencyData represents the storage and reference statistics for a given file tree path.
type EfficiencyData struct {
	Path              string
	Nodes             []*FileNode
	CumulativeSize    int64
	minDiscoveredSize int64
}

// EfficiencySlice represents an ordered set of EfficiencyData data structures.
type EfficiencySlice []*EfficiencyData
