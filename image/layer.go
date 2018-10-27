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

// Layer represents a Docker image layer and metadata
type Layer struct {
	TarPath  string
	History  ImageHistoryEntry
	Index    int
	Tree     *filetree.FileTree
	RefTrees []*filetree.FileTree
}

// ShortId returns the truncated id of the current layer.
func (layer *Layer) TarId() string {
	return strings.TrimSuffix(layer.TarPath, "/layer.tar")
}

// ShortId returns the truncated id of the current layer.
func (layer *Layer) Id() string {
	return layer.History.ID
}

// ShortId returns the truncated id of the current layer.
func (layer *Layer) ShortId() string {
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
func (layer *Layer) String() string {

	return fmt.Sprintf(LayerFormat,
		layer.ShortId(),
		humanize.Bytes(uint64(layer.History.Size)),
		strings.TrimPrefix(layer.History.CreatedBy, "/bin/sh -c "))
}
