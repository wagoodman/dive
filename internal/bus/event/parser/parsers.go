package parser

import (
	"fmt"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/internal/bus/event"
	"github.com/wagoodman/dive/internal/bus/event/payload"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"
)

type ErrBadPayload struct {
	Type  partybus.EventType
	Field string
	Value interface{}
}

func (e *ErrBadPayload) Error() string {
	return fmt.Sprintf("event='%s' has bad event payload field=%q: %q", string(e.Type), e.Field, e.Value)
}

func newPayloadErr(t partybus.EventType, field string, value interface{}) error {
	return &ErrBadPayload{
		Type:  t,
		Field: field,
		Value: value,
	}
}

func checkEventType(actual, expected partybus.EventType) error {
	if actual != expected {
		return newPayloadErr(expected, "Type", actual)
	}
	return nil
}

func ParseTaskStarted(e partybus.Event) (progress.StagedProgressable, *payload.GenericTask, error) {
	if err := checkEventType(e.Type, event.TaskStarted); err != nil {
		return nil, nil, err
	}

	var mon progress.StagedProgressable

	source, ok := e.Source.(payload.GenericTask)
	if !ok {
		return nil, nil, newPayloadErr(e.Type, "Source", e.Source)
	}

	mon, ok = e.Value.(progress.StagedProgressable)
	if !ok {
		mon = nil
	}

	return mon, &source, nil
}

// ParseExploreAnalysis returns the analysis to explore and the func to call once the UI is done with it (never nil).
func ParseExploreAnalysis(e partybus.Event) (*dive.Analysis, func(), error) {
	if err := checkEventType(e.Type, event.ExploreAnalysis); err != nil {
		return nil, func() {}, err
	}

	ex, ok := e.Value.(payload.Explore)
	if !ok {
		return nil, func() {}, newPayloadErr(e.Type, "Value", e.Value)
	}
	done := ex.Done
	if done == nil {
		done = func() {}
	}
	if ex.Analysis == nil {
		return nil, done, newPayloadErr(e.Type, "Value", e.Value)
	}

	return ex.Analysis, done, nil
}

func ParseReport(e partybus.Event) (string, string, error) {
	if err := checkEventType(e.Type, event.Report); err != nil {
		return "", "", err
	}

	context, ok := e.Source.(string)
	if !ok {
		// this is optional
		context = ""
	}

	report, ok := e.Value.(string)
	if !ok {
		return "", "", newPayloadErr(e.Type, "Value", e.Value)
	}

	return context, report, nil
}

func ParseNotification(e partybus.Event) (string, string, error) {
	if err := checkEventType(e.Type, event.Notification); err != nil {
		return "", "", err
	}

	context, ok := e.Source.(string)
	if !ok {
		// this is optional
		context = ""
	}

	notification, ok := e.Value.(string)
	if !ok {
		return "", "", newPayloadErr(e.Type, "Value", e.Value)
	}

	return context, notification, nil
}
