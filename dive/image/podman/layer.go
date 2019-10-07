package podman

import (
	podmanImage "github.com/containers/libpod/libpod/image"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"strings"
)

// Layer represents a Docker image layer and metadata
type layer struct {
	obj       *podmanImage.Image
	index     int
	tree      *filetree.FileTree
}

// ShortId returns the truncated id of the current layer.
func (l *layer) Command() string {
	if len(l.obj.ImageData.History) > 0 {
		hist := l.obj.ImageData.History
		return strings.TrimPrefix(hist[len(hist)-1].CreatedBy, "/bin/sh -c ")
	}
	return "unknown"
}

// ShortId returns the truncated id of the current layer.
func (l *layer) ShortId() string {
	rangeBound := 15
	id := l.obj.ID()
	if length := len(id); length < 15 {
		rangeBound = length
	}
	id = id[0:rangeBound]

	return id
}

// String represents a layer in a columnar format.
func (l *layer) ToLayer() *image.Layer {
	return &image.Layer{
		Id:      l.obj.ID(),
		Index:   l.index,
		Command: l.Command(),
		Size:    uint64(l.obj.ImageData.Size),
		Tree:    l.tree,
	}
}
