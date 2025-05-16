package adapter

import (
	"context"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/wagoodman/dive/dive/v1/image"
	"github.com/wagoodman/dive/internal/bus"
	"github.com/wagoodman/dive/internal/bus/event/payload"
	"github.com/wagoodman/dive/internal/log"
)

type Analyzer interface {
	Analyze(ctx context.Context, img *image.Image) (*image.Analysis, error)
}

type analysisActionObserver struct {
	Analyzer func(context.Context, *image.Image) (*image.Analysis, error)
}

func NewAnalyzer() Analyzer {
	return analysisActionObserver{
		Analyzer: image.Analyze,
	}
}

func (a analysisActionObserver) Analyze(ctx context.Context, img *image.Image) (*image.Analysis, error) {
	log.WithFields("image", img.Request).Infof("analyzing")

	layers := len(img.Layers)
	var files int
	var fileSize uint64
	for _, layer := range img.Layers {
		files += layer.Tree.Size
		fileSize += layer.Tree.FileSize
	}
	fileSizeStr := humanize.Bytes(fileSize)
	filesStr := humanize.Comma(int64(files))

	log.Debugf("├── layers: %d", layers)
	log.Debugf("├── files: %s", filesStr)
	log.Debugf("└── file size: %s", fileSizeStr)

	mon := bus.StartTask(payload.GenericTask{
		Title: payload.Title{
			Default:      "Analyzing image",
			WhileRunning: "Analyzing image",
			OnSuccess:    "Analyzed image",
		},
		HideOnSuccess:      false,
		HideStageOnSuccess: false,
		ID:                 img.Request,
		Context:            fmt.Sprintf("[layers:%d files:%s size:%s]", layers, filesStr, fileSizeStr),
	})

	analysis, err := a.Analyzer(ctx, img)
	if err != nil {
		mon.SetError(err)
	} else {
		mon.SetCompleted()
	}

	if err == nil && analysis == nil {
		err = fmt.Errorf("no results returned")
	}

	return analysis, err
}
