package handler

import (
	"errors"
	"strings"
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
		return payload.Title{Default: name, WhileRunning: name + "ing", OnSuccess: name + "ed", OnFail: "Failed to " + name}
	}
	failed := func() progress.StagedProgressable {
		// how dive's commands finish a failed task
		p := &payload.GenericProgress{
			AtomicStage: progress.NewAtomicStage(""),
			Manual:      progress.NewManual(-1),
		}
		p.SetError(errors.New("boom"))
		p.SetTotal(1)
		p.Manual.Set(1)
		return p
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
			name: "failed task",
			events: []partybus.Event{
				{
					Type:   event.TaskStarted,
					Source: payload.GenericTask{Title: title("Evaluat"), ID: "img", Context: "[rules: 3]"},
					Value:  failed(),
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

			// the frame's view of the models, in the order they were added
			var models []tea.Model
			update := func(msg tea.Msg) {
				for i, m := range models {
					models[i], _ = m.Update(msg)
				}
			}
			for _, e := range tt.events {
				newModels, cmd := h.Handle(e)
				require.NotNil(t, cmd)
				if len(newModels) > 0 {
					require.Len(t, newModels, 1, "a root task gets its own tree")
					models = append(models, newModels...)
					continue
				}
				// a sub task row is a message for its root's tree
				update(cmd())
			}

			// an ID-less tick reaches every row
			update(taskprogress.TickMsg{Time: time.Now()})

			var views []string
			for _, m := range models {
				// the frame leaves out models with nothing to show (a hidden finished task)
				if v := m.View(); v != "" {
					views = append(views, v)
				}
			}
			got := strings.Join(views, "\n")
			t.Log(got)
			snaps.MatchSnapshot(t, got)
		})
	}
}
