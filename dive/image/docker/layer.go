package docker

import (
	"github.com/wagoodman/dive/dive/image"
	"strings"

	"github.com/wagoodman/dive/dive/filetree"
)

// Layer represents a Docker image layer and metadata
type layer struct {
	history historyEntry
	index   int
	tree    *filetree.FileTree
}


// String represents a layer in a columnar format.
func (l *layer) ToLayer() *image.Layer {
	return &image.Layer{
		Id:      l.history.ID,
		Index:   l.index,
		Command: strings.TrimPrefix(l.history.CreatedBy, "/bin/sh -c "),
		Size:    l.history.Size,
		Tree:    l.tree,
	}
}
