package image

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/wagoodman/dive/filetree"
	"strings"
)

const (
	LayerFormat = "%-25s %7s  %s"
)

// ShortId returns the truncated id of the current layer.
func (layer *dockerLayer) TarId() string {
	return strings.TrimSuffix(layer.tarPath, "/layer.tar")
}

// ShortId returns the truncated id of the current layer.
func (layer *dockerLayer) Id() string {
	return layer.history.ID
}

// index returns the relative position of the layer within the image.
func (layer *dockerLayer) Index() int {
	return layer.index
}

// Size returns the number of bytes that this image is.
func (layer *dockerLayer) Size() uint64 {
	return layer.history.Size
}

// Tree returns the file tree representing the current layer.
func (layer *dockerLayer) Tree() *filetree.FileTree {
	return layer.tree
}

// ShortId returns the truncated id of the current layer.
func (layer *dockerLayer) Command() string {
	return strings.TrimPrefix(layer.history.CreatedBy, "/bin/sh -c ")
}

// ShortId returns the truncated id of the current layer.
func (layer *dockerLayer) ShortId() string {
	rangeBound := 25
	id := layer.Id()
	if length := len(id); length < 25 {
		rangeBound = length
	}
	id = id[0:rangeBound]

	// show the tagged image as the last layer
	// if len(layer.History.Tags) > 0 {
	// 	id = "[" + strings.Join(layer.History.Tags, ",") + "]"
	// }

	return id
}

// String represents a layer in a columnar format.
func (layer *dockerLayer) String() string {

	if layer.index == 0 {
		return fmt.Sprintf(LayerFormat,
			layer.ShortId(),
			humanize.Bytes(layer.Size()),
			"FROM "+layer.ShortId())
	}
	return fmt.Sprintf(LayerFormat,
		layer.ShortId(),
		humanize.Bytes(layer.Size()),
		layer.Command())
}
