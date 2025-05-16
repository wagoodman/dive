package handler

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wagoodman/go-partybus"

	"github.com/anchore/bubbly"
	"github.com/anchore/bubbly/bubbles/taskprogress"
	stereoscopeEvent "github.com/anchore/stereoscope/pkg/event"
	diveEvent "github.com/wagoodman/dive/internal/bus/event"
)

var _ interface {
	bubbly.EventHandler
	bubbly.MessageListener
	bubbly.HandleWaiter
} = (*Handler)(nil)

type HandlerConfig struct {
	TitleWidth        int
	AdjustDefaultTask func(taskprogress.Model) taskprogress.Model
}

type Handler struct {
	WindowSize tea.WindowSizeMsg
	Running    *sync.WaitGroup
	Config     HandlerConfig

	bubbly.EventHandler

	onNewCatalogerTask *sync.Once
}

func DefaultHandlerConfig() HandlerConfig {
	return HandlerConfig{
		TitleWidth: 30,
	}
}

func New(cfg HandlerConfig) *Handler {
	d := bubbly.NewEventDispatcher()

	h := &Handler{
		EventHandler:       d,
		Running:            &sync.WaitGroup{},
		Config:             cfg,
		onNewCatalogerTask: &sync.Once{},
	}

	// register all supported event types with the respective handler functions
	d.AddHandlers(map[partybus.EventType]bubbly.EventHandlerFn{
		stereoscopeEvent.PullDockerImage:     simpleHandler(h.handlePullDockerImage),
		stereoscopeEvent.PullContainerdImage: simpleHandler(h.handlePullContainerdImage),
		stereoscopeEvent.ReadImage:           simpleHandler(h.handleReadImage),
		stereoscopeEvent.FetchImage:          simpleHandler(h.handleFetchImage),
		diveEvent.TaskStarted:                h.handleTaskStarted,
		// TODO!
		//diveEvent.ExploreAnalysis: h.handleExploreAnalysis,
	})

	return h
}

func simpleHandler(fn func(partybus.Event) []tea.Model) bubbly.EventHandlerFn {
	return func(e partybus.Event) ([]tea.Model, tea.Cmd) {
		return fn(e), nil
	}
}

func (m *Handler) OnMessage(msg tea.Msg) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.WindowSize = msg
	}
}

func (m *Handler) Wait() {
	m.Running.Wait()
}
