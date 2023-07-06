package image

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"

	"github.com/wagoodman/dive/dive/filetree"
)

const (
	LayerFormat = "%7s  %s"
)

type Layer struct {
	Id      string
	Index   int
	Command string
	Size    uint64
	Tree    *filetree.FileTree
	Names   []string
	Digest  string
}

func (l *Layer) ShortId() string {
	rangeBound := 15
	id := l.Id
	if length := len(id); length < 15 {
		rangeBound = length
	}
	id = id[0:rangeBound]

	return id
}

func (l *Layer) commandPreview() string {
	// Layers using heredocs can be multiple lines; rendering relies on
	// Layer.String to be a single line.
	return strings.Replace(l.Command, "\n", "↵", -1)
}

func (l *Layer) String() string {
	if l.Index == 0 {
		return fmt.Sprintf(LayerFormat,
			humanize.Bytes(l.Size),
			"FROM "+l.ShortId())
	}
	return fmt.Sprintf(LayerFormat,
		humanize.Bytes(l.Size),
		l.commandPreview())
}
