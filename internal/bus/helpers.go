package bus

import (
	"context"
	"sync"

	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/internal/bus/event"
	"github.com/wagoodman/dive/internal/bus/event/payload"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"
)

func Report(report string) {
	if len(report) == 0 {
		return
	}
	Publish(partybus.Event{
		Type:  event.Report,
		Value: report,
	})
}

func Notify(message string) {
	Publish(partybus.Event{
		Type:  event.Notification,
		Value: message,
	})
}

func StartTask(info payload.GenericTask) *payload.GenericProgress {
	t := &payload.GenericProgress{
		AtomicStage: progress.NewAtomicStage(""),
		Manual:      progress.NewManual(-1),
	}

	Publish(partybus.Event{
		Type:   event.TaskStarted,
		Source: info,
		Value:  progress.StagedProgressable(t),
	})

	return t
}

// ExploreAnalysis hands the analysis to the TUI and blocks until the UI is done with it (or ctx is done). The TUI
// reads file contents from the analysis image for diffs and extraction, so the caller can only clean up the image
// once this returns.
func ExploreAnalysis(ctx context.Context, analysis *dive.Analysis) {
	done := make(chan struct{})
	var once sync.Once
	Publish(partybus.Event{
		Type:  event.ExploreAnalysis,
		Value: payload.Explore{Analysis: analysis, Done: func() { once.Do(func() { close(done) }) }},
	})
	select {
	case <-done:
	case <-ctx.Done():
	}
}
