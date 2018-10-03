package image

import (
	"github.com/docker/docker/api/types/image"
	"github.com/wagoodman/dive/filetree"
	"strings"
	"fmt"
	"strconv"
	"github.com/dustin/go-humanize"
)

type Layer struct {
	TarPath  string
	History  image.HistoryResponseItem
	Index    int
	Tree     *filetree.FileTree
	RefTrees []*filetree.FileTree
}

func (layer *Layer) Id() string {
	rangeBound := 25
	if length := len(layer.History.ID); length < 25 {
		rangeBound = length
	}
	id := layer.History.ID[0:rangeBound]

	// show the tagged image as the last layer
	// if len(layer.History.Tags) > 0 {
	// 	id = "[" + strings.Join(layer.History.Tags, ",") + "]"
	// }

	return id
}

func (layer *Layer) String() string {

	return fmt.Sprintf(LayerFormat,
		layer.Id(),
		strconv.Itoa(int(100.0*filetree.EfficiencyScore(layer.RefTrees[:layer.Index+1]))) + "%",
		//"100%",
		humanize.Bytes(uint64(layer.History.Size)),
		strings.TrimPrefix(layer.History.CreatedBy, "/bin/sh -c "))
}

