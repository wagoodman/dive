package image

import (
	"github.com/wagoodman/dive/dive/filetree"
)

const (
	LayerFormat = "%7s  %s"
)

type Layer interface {
	Id() string
	ShortId() string
	Index() int
	Command() string
	Size() uint64
	Tree() *filetree.FileTree
	String() string
}
