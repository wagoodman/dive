package docker

import (
	"fmt"
	"github.com/wagoodman/dive/dive/image"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/wagoodman/dive/dive/filetree"
)

// Layer represents a Docker image layer and metadata
type dockerLayer struct {
	tarPath string
	history imageHistoryEntry
	index   int
	tree    *filetree.FileTree
}

type imageHistoryEntry struct {
	ID         string
	Size       uint64
	Created    string `json:"created"`
	Author     string `json:"author"`
	CreatedBy  string `json:"created_by"`
	EmptyLayer bool   `json:"empty_layer"`
}

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
	rangeBound := 15
	id := layer.Id()
	if length := len(id); length < 15 {
		rangeBound = length
	}
	id = id[0:rangeBound]

	// show the tagged image as the last layer
	// if len(layer.History.Tags) > 0 {
	// 	id = "[" + strings.Join(layer.History.Tags, ",") + "]"
	// }

	return id
}

func (layer *dockerLayer) StringFormat() string {
	return image.LayerFormat
}

// String represents a layer in a columnar format.
func (layer *dockerLayer) String() string {

	if layer.index == 0 {
		return fmt.Sprintf(image.LayerFormat,
			// layer.ShortId(),
			// fmt.Sprintf("%d",layer.Index()),
			humanize.Bytes(layer.Size()),
			"FROM "+layer.ShortId())
	}
	return fmt.Sprintf(image.LayerFormat,
		// layer.ShortId(),
		// fmt.Sprintf("%d",layer.Index()),
		humanize.Bytes(layer.Size()),
		layer.Command())
}
