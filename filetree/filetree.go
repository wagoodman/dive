package filetree

import (
	"strings"
)

// FileTree represents a filesystem tree structure
type FileTree struct {
	Root *FileNode
}

// NewFileTree creates a new file tree
func NewFileTree() *FileTree {
	return &FileTree{
		Root: NewFileNode(nil, nil, "", FileInfo{}),
	}
}

// AddFile adds a file to the tree
func (tree *FileTree) AddFile(fileInfo FileInfo) {
	if tree.Root == nil {
		tree.Root = NewFileNode(tree, nil, "", FileInfo{})
	}

	parts := splitPath(fileInfo.Path)
	tree.Root.AddChild(parts, fileInfo)
}

// AddLink adds a symbolic or hard link to the tree
func (tree *FileTree) AddLink(name, linkName string) {
	if tree.Root == nil {
		tree.Root = NewFileNode(tree, nil, "", FileInfo{})
	}

	parts := splitPath(name)
	linkInfo := FileInfo{
		Path:     name,
		LinkName: linkName,
	}
	tree.Root.AddChild(parts, linkInfo)
}

// splitPath splits a file path into its components
func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}
	return strings.Split(strings.TrimPrefix(path, "/"), "/")
}
