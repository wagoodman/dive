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
	stereoscopeEvent "github.com/anchore/stereoscope/pkg/event"
)

func TestHandler_handleFetchImage(t *testing.T) {

	tests := []struct {
		name       string
		eventFn    func(*testing.T) partybus.Event
		iterations int
	}{
		{
			name: "fetch image in progress",
			eventFn: func(t *testing.T) partybus.Event {
				prog := &progress.Manual{}
				prog.SetTotal(100)
				prog.Set(50)

				mon := struct {
					progress.Progressable
					progress.Stager
				}{
					Progressable: prog,
					Stager: &progress.Stage{
						Current: "current",
					},
				}

				return partybus.Event{
					Type:   stereoscopeEvent.FetchImage,
					Source: "the-image",
					Value:  mon,
				}
			},
		},
		{
			name: "fetch image complete",
			eventFn: func(t *testing.T) partybus.Event {
				prog := &progress.Manual{}
				prog.SetTotal(100)
				prog.Set(100)
				prog.SetCompleted()

				mon := struct {
					progress.Progressable
					progress.Stager
				}{
					Progressable: prog,
					Stager: &progress.Stage{
						Current: "current",
					},
				}

				return partybus.Event{
					Type:   stereoscopeEvent.FetchImage,
					Source: "the-image",
					Value:  mon,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := tt.eventFn(t)
			handler := New(DefaultHandlerConfig())
			handler.WindowSize = tea.WindowSizeMsg{
				Width:  100,
				Height: 80,
			}

			models, _ := handler.Handle(event)
			require.Len(t, models, 1)
			model := models[0]

			tsk, ok := model.(taskprogress.Model)
			require.True(t, ok)

			gotModel := runModel(t, tsk, tt.iterations, taskprogress.TickMsg{
				Time:     time.Now(),
				Sequence: tsk.Sequence(),
				ID:       tsk.ID(),
			})

			got := gotModel.View()

			t.Log(got)
			snaps.MatchSnapshot(t, got)
		})
	}
}
