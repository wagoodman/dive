// copied and adapted from syft v1.55.0 (github.com/anchore/syft, cmd/syft/cli/ui)

package handler

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wagoodman/go-partybus"

	"github.com/anchore/bubbly/bubbles/taskprogress"
	stereoEventParsers "github.com/anchore/stereoscope/pkg/event/parsers"
	"github.com/wagoodman/dive/internal/log"
)

func (m *Handler) handleFetchImage(e partybus.Event) []tea.Model {
	imgName, prog, err := stereoEventParsers.ParseFetchImage(e)
	if err != nil {
		log.WithFields("error", err).Debug("unable to parse event")
		return nil
	}

	tsk := m.newTaskProgress(
		taskprogress.Title{
			Default: "Load image",
			Running: "Loading image",
			Success: "Loaded image",
		},
		taskprogress.WithStagedProgressable(prog),
	)
	if imgName != "" {
		tsk.Context = []string{imgName}
	}

	return []tea.Model{tsk}
}
