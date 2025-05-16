package bus

import (
	"github.com/anchore/clio"
	"github.com/wagoodman/dive/dive/v1/image"
	"github.com/wagoodman/dive/internal/bus/event"
	"github.com/wagoodman/dive/internal/bus/event/payload"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"
)

func Exit() {
	Publish(clio.ExitEvent(false))
}

func ExitWithInterrupt() {
	Publish(clio.ExitEvent(true))
}

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

func StartSizedTask(info payload.GenericTask, size int64, initialStage string) *payload.GenericProgress {
	t := &payload.GenericProgress{
		AtomicStage: progress.NewAtomicStage(initialStage),
		Manual:      progress.NewManual(size),
	}

	Publish(partybus.Event{
		Type:   event.TaskStarted,
		Source: info,
		Value:  progress.StagedProgressable(t),
	})

	return t
}

func ExploreAnalysis(analysis image.Analysis, reader image.ContentReader) {
	Publish(partybus.Event{
		Type:  event.ExploreAnalysis,
		Value: payload.Explore{Analysis: analysis, Content: reader},
	})
}
