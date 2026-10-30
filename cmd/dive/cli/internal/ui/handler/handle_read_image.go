// copied and adapted from syft v1.55.0 (github.com/anchore/syft, cmd/syft/cli/ui), the per layer rows are dive's own

package handler

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"

	"github.com/anchore/bubbly/bubbles/taskprogress"
	"github.com/anchore/bubbly/bubbles/tree"
	stereoEventParsers "github.com/anchore/stereoscope/pkg/event/parsers"
	"github.com/wagoodman/dive/internal/log"
)

const readImageRowID = "read-image"

var _ tea.Model = (*readImageModel)(nil)

// readImageModel shows the "parse image" row with one nested row per layer being indexed (from ReadLayer events)
type readImageModel struct {
	model tree.Model
	// the tree doesn't init the models it holds, so the read image row is started from here
	rootInit tea.Cmd
}

// newLayerRowMsg carries a layer row to the read image model. It is a message (not a direct mutation) so the tree
// is only changed inside the bubbletea update loop.
type newLayerRowMsg struct {
	id  string
	tsk taskprogress.Model
}

func (r readImageModel) Init() tea.Cmd {
	return r.rootInit
}

func (r readImageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if row, ok := msg.(newLayerRowMsg); ok {
		if err := r.model.Add(readImageRowID, row.id, row.tsk); err != nil {
			log.WithFields("error", err).Debug("unable to add layer row")
			return r, nil
		}
		return r, row.tsk.Init()
	}
	model, cmd := r.model.Update(msg)
	r.model = model.(tree.Model)
	return r, cmd
}

func (r readImageModel) View() string {
	return r.model.View()
}

func (m *Handler) handleReadImage(e partybus.Event) []tea.Model {
	imgMetadata, prog, err := stereoEventParsers.ParseReadImage(e)
	if err != nil {
		log.WithFields("error", err).Debug("unable to parse event")
		return nil
	}

	tsk := m.newTaskProgress(
		taskprogress.Title{
			Default: "Parse image",
			Running: "Parsing image",
			Success: "Parsed image",
		},
		taskprogress.WithProgress(prog),
	)

	if imgMetadata != nil {
		tsk.Context = []string{imgMetadata.ID}
	}

	t := tree.NewModel()
	t.Padding = "   "
	t.RootsWithoutPrefix = true
	if err := t.Add("", readImageRowID, tsk); err != nil {
		log.WithFields("error", err).Debug("unable to add read image row")
		return nil
	}

	return []tea.Model{readImageModel{model: t, rootInit: tsk.Init()}}
}

func (m *Handler) handleReadLayer(e partybus.Event) ([]tea.Model, tea.Cmd) {
	layer, mon, err := stereoEventParsers.ParseReadLayer(e)
	if err != nil {
		log.WithFields("error", err).Debug("unable to parse event")
		return nil, nil
	}

	// the digest goes in the title, not the right aligned context: the context is sized to the full window width,
	// which overflows once the tree indents the row
	title := fmt.Sprintf("layer %d", layer.Index)
	if d := strings.TrimPrefix(layer.Digest, "sha256:"); d != "" {
		title += " " + d[:min(len(d), 12)]
	}
	tsk := m.newTaskProgress(
		taskprogress.Title{
			Default: title,
			Running: title,
			Success: title,
		},
		taskprogress.WithStagedProgressable(layerReadProgress{Monitorable: mon}),
	)
	tsk.TitleStyle = lipgloss.NewStyle()
	// finished layers drop out so a many layer image doesn't flood the screen, the parent row has the overall progress
	tsk.HideOnSuccess = true

	id := fmt.Sprintf("layer-%d", layer.Index)
	return nil, func() tea.Msg {
		return newLayerRowMsg{id: id, tsk: tsk}
	}
}

var _ progress.StagedProgressable = (*layerReadProgress)(nil)

// layerReadProgress adapts the ReadLayer monitor (a count of indexed tar entries with no known total)
type layerReadProgress struct {
	progress.Monitorable
}

func (l layerReadProgress) Size() int64 {
	return -1
}

func (l layerReadProgress) Stage() string {
	return humanize.Comma(l.Current()) + " files"
}
