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

// we standardize how rows are instantiated to ensure consistency in the appearance across the UI
type taskModelFactory func(title taskprogress.Title, opts ...taskprogress.Option) taskprogress.Model

var _ tea.Model = (*taskTreeModel)(nil)

// taskTreeModel holds every dive task row (load, analyze, evaluate, export...) in one tree
type taskTreeModel struct {
	model        tree.Model
	modelFactory taskModelFactory
}

func newTaskTreeModel(f taskModelFactory) *taskTreeModel {
	t := tree.NewModel()
	t.Padding = "   "
	t.RootsWithoutPrefix = true
	return &taskTreeModel{
		modelFactory: f,
		model:        t,
	}
}

type newTaskRowEvent struct {
	info payload.GenericTask
	prog progress.StagedProgressable
}

func (cts taskTreeModel) Init() tea.Cmd {
	return cts.model.Init()
}

func (cts taskTreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	event, ok := msg.(newTaskRowEvent)
	if !ok {
		model, cmd := cts.model.Update(msg)
		cts.model = model.(tree.Model)

		return cts, cmd
	}

	info, prog := event.info, event.prog

	opts := []taskprogress.Option{}
	if prog != nil {
		opts = append(opts, taskprogress.WithStagedProgressable(prog))
	}

	tsk := cts.modelFactory(
		taskprogress.Title{
			Default: info.Title.Default,
			Running: info.Title.WhileRunning,
			Success: info.Title.OnSuccess,
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

	if err := cts.model.Add(info.ParentID, info.ID, tsk); err != nil {
		log.WithFields("error", err).Error("unable to add task to tree model")
	}

	return cts, tsk.Init()
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

	var models []tea.Model

	// only create the task tree once to manage all task events
	m.onNewTask.Do(func() {
		models = append(models, newTaskTreeModel(m.newTaskProgress))
	})

	// we should never update a model outside of the bubbletea update-render loop, so the new row is sent as a message
	// (produced by a command) that the task tree model picks up.

	// the ID is optional from the consumer perspective, but required (and unique) internally. dive reuses the image
	// name as the ID for several unrelated tasks, so a repeat gets a fresh ID instead of replacing the earlier row.
	if info.ID == "" || m.taskIDs[info.ID] {
		info.ID = uuid.Must(uuid.NewRandom()).String()
	}
	m.taskIDs[info.ID] = true

	cmd := func() tea.Msg {
		return newTaskRowEvent{
			info: *info,
			prog: mon,
		}
	}

	return models, cmd
}
