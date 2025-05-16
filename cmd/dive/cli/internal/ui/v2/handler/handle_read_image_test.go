package handler

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gkampitakis/go-snaps/snaps"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/stretchr/testify/require"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"

	"github.com/anchore/bubbly/bubbles/taskprogress"
	stereoscopeEvent "github.com/anchore/stereoscope/pkg/event"
	"github.com/anchore/stereoscope/pkg/image"
)

func TestHandler_handleReadImage(t *testing.T) {

	tests := []struct {
		name       string
		eventFn    func(*testing.T) partybus.Event
		iterations int
	}{
		{
			name: "read image in progress",
			eventFn: func(t *testing.T) partybus.Event {
				prog := &progress.Manual{}
				prog.SetTotal(100)
				prog.Set(50)

				src := image.Metadata{
					ID:   "id",
					Size: 42,
					Config: v1.ConfigFile{
						Architecture: "arch",
						Author:       "auth",
						Container:    "cont",
						OS:           "os",
						OSVersion:    "os-ver",
						Variant:      "vari",
					},
					MediaType:      "media",
					ManifestDigest: "digest",
					Architecture:   "arch",
					Variant:        "var",
					OS:             "os",
				}

				return partybus.Event{
					Type:   stereoscopeEvent.ReadImage,
					Source: src,
					Value:  prog,
				}
			},
		},
		{
			name: "read image complete",
			eventFn: func(t *testing.T) partybus.Event {
				prog := &progress.Manual{}
				prog.SetTotal(100)
				prog.Set(100)
				prog.SetCompleted()

				src := image.Metadata{
					ID:   "id",
					Size: 42,
					Config: v1.ConfigFile{
						Architecture: "arch",
						Author:       "auth",
						Container:    "cont",
						OS:           "os",
						OSVersion:    "os-ver",
						Variant:      "vari",
					},
					MediaType:      "media",
					ManifestDigest: "digest",
					Architecture:   "arch",
					Variant:        "var",
					OS:             "os",
				}

				return partybus.Event{
					Type:   stereoscopeEvent.ReadImage,
					Source: src,
					Value:  prog,
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
