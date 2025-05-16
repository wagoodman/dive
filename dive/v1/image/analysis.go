package image

import (
	"context"
	"github.com/wagoodman/dive/dive/v1/filetree"
)

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

func Analyze(ctx context.Context, img *Image) (*Analysis, error) {
	efficiency, inefficiencies := filetree.Efficiency(img.Trees)
	var sizeBytes, userSizeBytes uint64

	for i, v := range img.Layers {
		sizeBytes += v.Size
		if i != 0 {
			userSizeBytes += v.Size
		}
	}

	var wastedBytes uint64
	for _, file := range inefficiencies {
		wastedBytes += uint64(file.CumulativeSize)
	}

	return &Analysis{
		Image:             img.Request,
		Layers:            img.Layers,
		RefTrees:          img.Trees,
		Efficiency:        efficiency,
		UserSizeByes:      userSizeBytes,
		SizeBytes:         sizeBytes,
		WastedBytes:       wastedBytes,
		WastedUserPercent: float64(wastedBytes) / float64(userSizeBytes),
		Inefficiencies:    inefficiencies,
	}, nil
}
