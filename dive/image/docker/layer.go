package docker

import (
	"fmt"
	"github.com/wagoodman/dive/dive/image"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/wagoodman/dive/dive/filetree"
)

// Layer represents a Docker image layer and metadata
type layer struct {
	history historyEntry
	index   int
	tree    *filetree.FileTree
}

// ShortId returns the truncated id of the current layer.
func (l *layer) Id() string {
	return l.history.ID
}

// index returns the relative position of the layer within the image.
func (l *layer) Index() int {
	return l.index
}

// Size returns the number of bytes that this image is.
func (l *layer) Size() uint64 {
	return l.history.Size
}

// Tree returns the file tree representing the current layer.
func (l *layer) Tree() *filetree.FileTree {
	return l.tree
}

// ShortId returns the truncated id of the current layer.
func (l *layer) Command() string {
	return strings.TrimPrefix(l.history.CreatedBy, "/bin/sh -c ")
}

// ShortId returns the truncated id of the current layer.
func (l *layer) ShortId() string {
	rangeBound := 15
	id := l.Id()
	if length := len(id); length < 15 {
		rangeBound = length
	}
	id = id[0:rangeBound]

	return id
}

// String represents a layer in a columnar format.
func (l *layer) String() string {

	if l.index == 0 {
		return fmt.Sprintf(image.LayerFormat,
			humanize.Bytes(l.Size()),
			"FROM "+l.ShortId())
	}
	return fmt.Sprintf(image.LayerFormat,
		humanize.Bytes(l.Size()),
		l.Command())
}
