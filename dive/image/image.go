package image

import (
	"github.com/wagoodman/dive/dive/filetree"
)

type Image struct {
	Request string
	Trees   []*filetree.FileTree
	Layers  []*Layer
}
