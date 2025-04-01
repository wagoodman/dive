package image

import (
	"github.com/wagoodman/dive/dive/filetree"
	"golang.org/x/net/context"
)

type Image struct {
	Request string
	Trees   []*filetree.FileTree
	Layers  []*Layer
}

func (img *Image) Analyze(ctx context.Context) (*Analysis, error) {
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
