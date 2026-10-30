// copied and adapted from syft v1.55.0 (github.com/anchore/syft, cmd/syft/cli/ui/handle_cataloger_task.go)

package handler

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"

	"github.com/anchore/bubbly/bubbles/taskprogress"
	"github.com/anchore/bubbly/bubbles/tree"
	"github.com/wagoodman/dive/internal/bus/event/parser"
	"github.com/wagoodman/dive/internal/bus/event/payload"
	"github.com/wagoodman/dive/internal/log"
)

var _ tea.Model = (*taskTreeModel)(nil)

// taskTreeModel holds one root dive task row (load, analyze, evaluate, export...) and its sub tasks. Each root gets
// its own model, so rows render in the order their tasks started, between the stereoscope rows.
type taskTreeModel struct {
	id    string
	model tree.Model
}

func newTaskTreeModel(id string) *taskTreeModel {
	t := tree.NewModel()
	t.Padding = "   "
	t.RootsWithoutPrefix = true
	return &taskTreeModel{
		id:    id,
		model: t,
	}
}

// newTaskRowEvent carries a sub task row to the tree of its root task
type newTaskRowEvent struct {
	treeID, parentID, id string
	tsk                  taskprogress.Model
}

func (cts taskTreeModel) Init() tea.Cmd {
	return cts.model.Init()
}

func (cts taskTreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	row, ok := msg.(newTaskRowEvent)
	if !ok {
		model, cmd := cts.model.Update(msg)
		cts.model = model.(tree.Model)

		return cts, cmd
	}
	if row.treeID != cts.id {
		return cts, nil
	}
	if err := cts.model.Add(row.parentID, row.id, row.tsk); err != nil {
		log.WithFields("error", err).Error("unable to add task to tree model")
		return cts, nil
	}
	return cts, row.tsk.Init()
}

func (cts taskTreeModel) View() string {
	return cts.model.View()
}

func (m *Handler) handleTaskStarted(e partybus.Event) ([]tea.Model, tea.Cmd) {
	mon, info, err := parser.ParseTaskStarted(e)
	if err != nil {
		log.WithFields("error", err).Debug("unable to parse event")
		return nil, nil
	}

	// the ID is optional from the consumer perspective, but required (and unique) internally. dive reuses the image
	// name as the ID for several unrelated tasks, so a repeat gets a fresh ID instead of replacing the earlier row.
	if info.ID == "" || m.taskIDs[info.ID] != "" {
		info.ID = uuid.Must(uuid.NewRandom()).String()
	}

	// the row is built now, while the event is handled, so the handler's wait group counts it before anyone waits
	tsk := m.newTaskRow(*info, mon)

	treeID, isSub := m.taskIDs[info.ParentID]
	if info.ParentID == "" || !isSub {
		m.taskIDs[info.ID] = info.ID
		t := newTaskTreeModel(info.ID)
		if err := t.model.Add("", info.ID, tsk); err != nil {
			log.WithFields("error", err).Error("unable to add task to tree model")
			return nil, nil
		}
		return []tea.Model{t}, tsk.Init()
	}

	// we should never update a model outside of the bubbletea update-render loop, so a sub task row is sent as a
	// message (produced by a command) that its root's tree model picks up
	m.taskIDs[info.ID] = treeID
	row := newTaskRowEvent{treeID: treeID, parentID: info.ParentID, id: info.ID, tsk: tsk}
	return nil, func() tea.Msg { return row }
}

func (m *Handler) newTaskRow(info payload.GenericTask, prog progress.StagedProgressable) taskprogress.Model {
	var opts []taskprogress.Option
	if prog != nil {
		opts = append(opts, taskprogress.WithStagedProgressable(prog))
	}

	tsk := m.newTaskProgress(
		taskprogress.Title{
			Default: info.Title.Default,
			Running: info.Title.WhileRunning,
			Success: info.Title.OnSuccess,
			Failed:  info.Title.OnFail,
		},
		opts...,
	)

	if info.Context != "" {
		tsk.Context = []string{info.Context}
	}

	tsk.HideOnSuccess = info.HideOnSuccess
	tsk.HideStageOnSuccess = info.HideStageOnSuccess
	tsk.HideProgressOnSuccess = true

	if info.ParentID != "" {
		tsk.TitleStyle = lipgloss.NewStyle()
	}
	return tsk
}
