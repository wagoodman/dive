package image

import (
	"github.com/wagoodman/dive/dive/v1/filetree"
)

type Image struct {
	Request string
	Trees   []*filetree.FileTree
	Layers  []*Layer
}
