package viewmodel

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"

	"github.com/wagoodman/dive/dive"
)

const LayerFormat = "%7s  %s"

// Layer is what the layer views show about a dive.Layer
type Layer struct {
	dive.Layer
	// Id is the layer digest without the algorithm prefix
	Id    string
	Names []string
}

func NewLayers(layers []dive.Layer) []*Layer {
	out := make([]*Layer, len(layers))
	for i, l := range layers {
		_, id, found := strings.Cut(l.Digest, ":")
		if !found {
			id = l.Digest
		}
		out[i] = &Layer{
			Layer: l,
			Id:    id,
			// todo: query the image source for tags
			Names: []string{"(unavailable)"},
		}
	}
	return out
}

func (l *Layer) ShortId() string {
	return l.Id[:min(15, len(l.Id))]
}

func (l *Layer) commandPreview() string {
	// layers using heredocs can be multiple lines, rendering relies on String() being a single line
	return strings.ReplaceAll(l.Command, "\n", "↵")
}

func (l *Layer) String() string {
	if l.Index == 0 {
		return fmt.Sprintf(LayerFormat, humanize.Bytes(l.Size), "FROM "+l.ShortId())
	}
	return fmt.Sprintf(LayerFormat, humanize.Bytes(l.Size), l.commandPreview())
}
