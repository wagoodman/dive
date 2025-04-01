package image

import (
	"github.com/wagoodman/dive/dive/filetree"
	"golang.org/x/net/context"
)

type Analyzer interface {
	Analyze(context.Context) (*Analysis, error)
}

type Analysis struct {
	Image             string
	Layers            []*Layer
	RefTrees          []*filetree.FileTree
	Efficiency        float64
	SizeBytes         uint64
	UserSizeByes      uint64  // this is all bytes except for the base image
	WastedUserPercent float64 // = wasted-bytes/user-size-bytes
	WastedBytes       uint64
	Inefficiencies    filetree.EfficiencySlice
}
