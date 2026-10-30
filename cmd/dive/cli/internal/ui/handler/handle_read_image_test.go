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
	"github.com/anchore/stereoscope/pkg/image"
)

func TestHandler_handleReadImage(t *testing.T) {
	type layer struct {
		files    int64
		complete bool
	}
	tests := []struct {
		name      string
		imageDone bool
		layers    []layer
	}{
		{
			name: "read image with layers in progress",
			layers: []layer{
				{files: 120, complete: true},
				{files: 1234},
				{files: 0},
			},
		},
		{
			name:      "read image complete",
			imageDone: true,
			layers: []layer{
				{files: 120, complete: true},
				{files: 1234, complete: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(DefaultHandlerConfig())
			h.WindowSize = tea.WindowSizeMsg{Width: 120, Height: 80}

			prog := progress.NewManual(int64(len(tt.layers) * 2))
			prog.Set(int64(len(tt.layers)))
			if tt.imageDone {
				prog.SetCompleted()
			}

			models, _ := h.Handle(partybus.Event{
				Type:   stereoscopeEvent.ReadImage,
				Source: image.Metadata{ID: "sha256:image-id"},
				Value:  progress.Progressable(prog),
			})
			require.Len(t, models, 1)
			model := models[0]
			model.Init()

			for idx, l := range tt.layers {
				mon := &progress.Manual{}
				mon.Set(l.files)
				if l.complete {
					mon.SetCompleted()
				}
				newModels, cmd := h.Handle(partybus.Event{
					Type:   stereoscopeEvent.ReadLayer,
					Source: image.LayerMetadata{Index: uint(idx), Digest: "sha256:0123456789abcdef0123"},
					Value:  progress.Monitorable(mon),
				})
				require.Empty(t, newModels)
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
