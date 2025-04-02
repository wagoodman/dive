package adapter

import (
	"context"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/internal/bus"
	"github.com/wagoodman/dive/internal/bus/event/payload"
	"github.com/wagoodman/dive/internal/log"
	"strings"
	"time"
)

type imageActionObserver struct {
	image.Resolver
}

func ImageResolver(resolver image.Resolver) image.Resolver {
	return imageActionObserver{
		Resolver: resolver,
	}
}

func (i imageActionObserver) Build(ctx context.Context, options []string) (*image.Image, error) {
	log.Info("building image")
	log.Debugf("└── %s", strings.Join(options, " "))

	mon := bus.StartTask(payload.GenericTask{
		Title: payload.Title{
			Default:      "Building image",
			WhileRunning: "Building image",
			OnSuccess:    "Built image",
		},
		HideOnSuccess:      false,
		HideStageOnSuccess: false,
		Context:            "... " + strings.Join(options, " "),
	})

	img, err := i.Resolver.Build(ctx, options)
	if err != nil {
		mon.SetError(err)
	} else {
		mon.SetCompleted()
	}
	return img, err
}

func (i imageActionObserver) Fetch(ctx context.Context, id string) (*image.Image, error) {
	log.WithFields("image", id).Info("fetching")
	log.Debugf("└── resolver: %s", i.Resolver.Name())

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mon := bus.StartTask(payload.GenericTask{
		Title: payload.Title{
			Default:      "Fetching image",
			WhileRunning: "Fetching image",
			OnSuccess:    "Fetched image",
		},
		HideOnSuccess:      false,
		HideStageOnSuccess: false,
		ID:                 id,
		Context:            id,
	})

	go func() {
		// in 5 seconds if the context is not cancelled, log the message
		select { // nolint:gosimple
		case <-time.After(3 * time.Second):
			if ctx.Err() == nil {
				bus.Notify(" • this can take a while for large images...")
				mon.AtomicStage.Set("(this can take a while for large images)")

				// TODO: default level should be error for this to work when using the UI
				//log.Warn("this can take a while for large images")
			}
		}
	}()

	img, err := i.Resolver.Fetch(ctx, id)
	if err != nil {
		mon.SetError(err)
	} else {
		mon.SetCompleted()
	}
	return img, err
}
