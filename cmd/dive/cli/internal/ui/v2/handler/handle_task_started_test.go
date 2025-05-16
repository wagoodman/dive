package handler

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"

	syftEvent "github.com/anchore/syft/syft/event"
	"github.com/anchore/syft/syft/event/monitor"
)

func TestHandler_handleCatalogerTaskStarted(t *testing.T) {
	title := monitor.Title{
		Default:   "some task title",
		OnSuccess: "some task done",
	}
	tests := []struct {
		name       string
		eventFn    func(*testing.T) partybus.Event
		iterations int
	}{
		{
			name: "cataloging task in progress",
			eventFn: func(t *testing.T) partybus.Event {
				value := &monitor.CatalogerTaskProgress{
					AtomicStage: progress.NewAtomicStage("some stage"),
					Manual:      progress.NewManual(100),
				}

				value.Manual.Add(50)

				return partybus.Event{
					Type: syftEvent.CatalogerTaskStarted,
					Source: monitor.GenericTask{
						Title:              title,
						HideOnSuccess:      false,
						HideStageOnSuccess: false,
						ID:                 "my-id",
					},
					Value: value,
				}
			},
		},
		{
			name: "cataloging sub task in progress",
			eventFn: func(t *testing.T) partybus.Event {
				value := &monitor.CatalogerTaskProgress{
					AtomicStage: progress.NewAtomicStage("some stage"),
					Manual:      progress.NewManual(100),
				}

				value.Manual.Add(50)

				return partybus.Event{
					Type: syftEvent.CatalogerTaskStarted,
					Source: monitor.GenericTask{
						Title:              title,
						HideOnSuccess:      false,
						HideStageOnSuccess: false,
						ID:                 "my-id",
						ParentID:           "top-level-task",
					},
					Value: value,
				}
			},
		},
		{
			name: "cataloging sub task complete",
			eventFn: func(t *testing.T) partybus.Event {
				value := &monitor.CatalogerTaskProgress{
					AtomicStage: progress.NewAtomicStage("some stage"),
					Manual:      progress.NewManual(100),
				}

				value.SetCompleted()

				return partybus.Event{
					Type: syftEvent.CatalogerTaskStarted,
					Source: monitor.GenericTask{
						Title:              title,
						HideOnSuccess:      false,
						HideStageOnSuccess: false,
						ID:                 "my-id",
						ParentID:           "top-level-task",
					},
					Value: value,
				}
			},
		},
		{
			name: "cataloging sub task complete -- hide stage",
			eventFn: func(t *testing.T) partybus.Event {
				value := &monitor.CatalogerTaskProgress{
					AtomicStage: progress.NewAtomicStage("some stage"),
					Manual:      progress.NewManual(100),
				}

				value.SetCompleted()

				return partybus.Event{
					Type: syftEvent.CatalogerTaskStarted,
					Source: monitor.GenericTask{
						Title:              title,
						HideOnSuccess:      false,
						HideStageOnSuccess: true,
						ID:                 "my-id",
						ParentID:           "top-level-task",
					},
					Value: value,
				}
			},
		},
		{
			name: "cataloging sub task complete with removal",
			eventFn: func(t *testing.T) partybus.Event {
				value := &monitor.CatalogerTaskProgress{
					AtomicStage: progress.NewAtomicStage("some stage"),
					Manual:      progress.NewManual(100),
				}

				value.SetCompleted()

				return partybus.Event{
					Type: syftEvent.CatalogerTaskStarted,
					Source: monitor.GenericTask{
						Title:              title,
						HideOnSuccess:      true,
						HideStageOnSuccess: false,
						ID:                 "my-id",
						ParentID:           "top-level-task",
					},
					Value: value,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// need to be able to get the initial newTaskRowEvent + initialize the nested taskprogress model
			if tt.iterations == 0 {
				tt.iterations = 2
			}

			e := tt.eventFn(t)
			handler := New(DefaultHandlerConfig())
			handler.WindowSize = tea.WindowSizeMsg{
				Width:  100,
				Height: 80,
			}

			info := monitor.GenericTask{
				Title: monitor.Title{
					Default:      "Catalog contents",
					WhileRunning: "Cataloging contents",
					OnSuccess:    "Cataloged contents",
				},
				ID: "top-level-task",
			}

			// note: this line / event is not under test, only needed to show a sub status
			kickoffEvent := &monitor.CatalogerTaskProgress{
				AtomicStage: progress.NewAtomicStage(""),
				Manual:      progress.NewManual(-1),
			}

			models, cmd := handler.Handle(
				partybus.Event{
					Type:   syftEvent.CatalogerTaskStarted,
					Source: info,
					Value:  progress.StagedProgressable(kickoffEvent),
				},
			)
			require.Len(t, models, 1)
			require.NotNil(t, cmd)
			model := models[0]

			tr, ok := model.(*taskModel)
			require.True(t, ok)

			gotModel := runModel(t, tr, tt.iterations, cmd())

			models, cmd = handler.Handle(e)
			require.Len(t, models, 0)
			require.NotNil(t, cmd)

			gotModel = runModel(t, gotModel, tt.iterations, cmd())

			got := gotModel.View()

			t.Log(got)
			snaps.MatchSnapshot(t, got)
		})
	}
}
