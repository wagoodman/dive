package viewmodel

import (
	"github.com/wagoodman/dive/dive/v1/image"
)

type LayerSelection struct {
	Layer                                                      *image.Layer
	BottomTreeStart, BottomTreeStop, TopTreeStart, TopTreeStop int
}
