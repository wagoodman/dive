package handler

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"

	"github.com/anchore/bubbly/bubbles/taskprogress"
	stereoscopeEvent "github.com/anchore/stereoscope/pkg/event"
	"github.com/anchore/stereoscope/pkg/image/docker"
	diveEvent "github.com/wagoodman/dive/internal/bus/event"
)

// tick renders a task row after one update
func tick(t *testing.T, m tea.Model) string {
	t.Helper()
	tsk, ok := m.(taskprogress.Model)
	require.True(t, ok)
	return runModel(t, tsk, 1, taskprogress.TickMsg{Time: time.Now(), Sequence: tsk.Sequence(), ID: tsk.ID()}).View()
}

func TestHandler_sourceFailedFailsPullRow(t *testing.T) {
	h := New(DefaultHandlerConfig())
	h.WindowSize = tea.WindowSizeMsg{Width: 100, Height: 80}

	models, _ := h.Handle(partybus.Event{Type: stereoscopeEvent.PullDockerImage, Source: "img", Value: &docker.PullStatus{}})
	require.Len(t, models, 1)

	pullErr := errors.New("pull access denied")
	_, _ = h.Handle(partybus.Event{Type: diveEvent.ImageSourceFailed, Source: "docker", Error: pullErr})
	view := tick(t, models[0])
	assert.Contains(t, view, "✘")
	assert.NotContains(t, view, "Pulled image")

	// stereoscope completes the pull status even when the pull failed, the failure has to win
	f := h.failable(newDockerPullProgressAdapter(mockDockerPullStatus{complete: true}))
	_, _ = h.Handle(partybus.Event{Type: diveEvent.ImageSourceFailed, Source: "docker", Error: pullErr})
	assert.True(t, progress.IsCompleted(f))
	assert.ErrorIs(t, f.Error(), pullErr)

	// rows of the next attempt are not failed by an earlier one
	next := h.failable(newDockerPullProgressAdapter(mockDockerPullStatus{complete: true}))
	assert.ErrorIs(t, next.Error(), progress.ErrCompleted)
}
