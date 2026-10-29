package handler

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"

	"github.com/anchore/bubbly/bubbles/taskprogress"
	"github.com/wagoodman/dive/internal/bus/event"
	"github.com/wagoodman/dive/internal/bus/event/payload"
)

func TestHandler_handleTaskStarted(t *testing.T) {
	newProg := func(stage string, done bool) progress.StagedProgressable {
		p := &payload.GenericProgress{
			AtomicStage: progress.NewAtomicStage(stage),
			Manual:      progress.NewManual(-1),
		}
		if done {
			p.SetCompleted()
		}
		return p
	}
	title := func(name string) payload.Title {
		return payload.Title{Default: name, WhileRunning: name + "ing", OnSuccess: name + "ed"}
	}

	tests := []struct {
		name   string
		events []partybus.Event
	}{
		{
			name: "task in progress",
			events: []partybus.Event{
				{
					Type:   event.TaskStarted,
					Source: payload.GenericTask{Title: title("Analyz"), ID: "img", Context: "[layers:3]"},
					Value:  newProg("some stage", false),
				},
			},
		},
		{
			name: "tasks sharing an ID all get rows",
			events: []partybus.Event{
				{
					Type:   event.TaskStarted,
					Source: payload.GenericTask{Title: title("Load"), ID: "img", Context: "img"},
					Value:  newProg("", true),
				},
				{
					Type:   event.TaskStarted,
					Source: payload.GenericTask{Title: title("Analyz"), ID: "img", Context: "[layers:3]"},
					Value:  newProg("", true),
				},
				{
					Type:   event.TaskStarted,
					Source: payload.GenericTask{Title: title("Evaluat"), ID: "img", Context: "[rules: 3]"},
					Value:  newProg("checking", false),
				},
			},
		},
		{
			name: "sub task and hidden task",
			events: []partybus.Event{
				{
					Type:   event.TaskStarted,
					Source: payload.GenericTask{Title: title("Build"), ID: "build"},
					Value:  newProg("step 2/3", false),
				},
				{
					Type:   event.TaskStarted,
					Source: payload.GenericTask{Title: title("Push"), ID: "push", ParentID: "build"},
					Value:  newProg("", false),
				},
				{
					Type:   event.TaskStarted,
					Source: payload.GenericTask{Title: title("Hid"), HideOnSuccess: true},
					Value:  newProg("", true),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(DefaultHandlerConfig())
			h.WindowSize = tea.WindowSizeMsg{Width: 100, Height: 80}

			var model tea.Model
			for _, e := range tt.events {
				models, cmd := h.Handle(e)
				if model == nil {
					require.Len(t, models, 1, "the first task creates the task tree")
					model = models[0]
				} else {
					require.Empty(t, models, "later tasks reuse the task tree")
				}
				require.NotNil(t, cmd)
				model, _ = model.Update(cmd())
			}

			// an ID-less tick reaches every row
			model, _ = model.Update(taskprogress.TickMsg{Time: time.Now()})

			got := model.View()
			t.Log(got)
			snaps.MatchSnapshot(t, got)
		})
	}
}
