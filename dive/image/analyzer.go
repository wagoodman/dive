package image

import (
	"github.com/buildpacks/lifecycle"
	"github.com/wagoodman/dive/dive/filetree"
)

type Analyzer interface {
	Analyze() (*AnalysisResult, error)
}

type AnalysisResult struct {
	ImageName string
	Layers            []*Layer
	RefTrees          []*filetree.FileTree
	Efficiency        float64
	SizeBytes         uint64
	UserSizeByes      uint64  // this is all bytes except for the base image
	WastedUserPercent float64 // = wasted-bytes/user-size-bytes
	WastedBytes       uint64
	Inefficiencies    filetree.EfficiencySlice
	BOMMapping		  map[string]lifecycle.BOMEntry
}


