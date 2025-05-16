package ui

import (
	"fmt"
	"github.com/anchore/bubbly"
	"github.com/anchore/bubbly/bubbles/frame"
	"github.com/anchore/clio"
	"github.com/anchore/go-logger"
	"github.com/anchore/go-logger/adapter/discard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/go-multierror"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v2/handler"
	"github.com/wagoodman/dive/internal/bus"
	"github.com/wagoodman/dive/internal/bus/event"
	"github.com/wagoodman/dive/internal/bus/event/parser"
	"github.com/wagoodman/dive/internal/log"
	"github.com/wagoodman/go-partybus"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

var _ clio.UI = (*V1UI)(nil)

type v2UI struct {
	// TODO!
	//cfg            v1.Preferences
	out            io.Writer
	err            io.Writer
	program        *tea.Program
	running        *sync.WaitGroup
	finalizeEvents []partybus.Event
	handler        *bubbly.HandlerCollection
	subscription   partybus.Unsubscribable
	quiet          bool
	verbosity      int
	format         v2Format
	frame          tea.Model
}

type v2Format struct {
	Title        lipgloss.Style
	Aux          lipgloss.Style
	Line         lipgloss.Style
	Notification lipgloss.Style
}

type v2BootstrapUI struct {
	app          clio.Application
	exploreUI    *v2UI
	subscription partybus.Unsubscribable
}

func NewV2UI(app clio.Application, out io.Writer, quiet bool, verbosity int) clio.UI {
	exploreUI := newV2UI(out, quiet, verbosity)
	if verbosity == 0 && !quiet {
		return exploreUI
	}
	// TODO: opts.V1Preferences(),
	return &v2BootstrapUI{
		app:       app,
		exploreUI: exploreUI,
	}
}

func (n *v2BootstrapUI) Setup(subscription partybus.Unsubscribable) error {
	n.subscription = subscription
	return nil
}

func (n *v2BootstrapUI) Handle(e partybus.Event) error {
	if e.Type == event.ExploreAnalysis {
		type Stater interface {
			State() *clio.State
		}

		state := n.app.(Stater).State()

		return state.UI.Replace(n.exploreUI)
	}
	return nil
}

func (n v2BootstrapUI) Teardown(_ bool) error {
	return nil
}

func newV2UI(out io.Writer, quiet bool, verbosity int) *v2UI {
	return &v2UI{
		//cfg:       cfg,
		out:       out,
		err:       os.Stderr,
		quiet:     quiet,
		verbosity: verbosity,
		running:   &sync.WaitGroup{},
		handler:   bubbly.NewHandlerCollection(handler.New(handler.DefaultHandlerConfig())),
		frame:     frame.New(),
		format: v2Format{
			Title:        lipgloss.NewStyle().Bold(true).Width(30),
			Aux:          lipgloss.NewStyle().Faint(true),
			Notification: lipgloss.NewStyle().Foreground(lipgloss.Color("#A77BCA")),
		},
	}
}

func (m v2UI) Init() tea.Cmd {
	return m.frame.Init()
}

func (m v2UI) RespondsTo() []partybus.EventType {
	return append([]partybus.EventType{
		event.Report,
		event.Notification,
	}, m.handler.RespondsTo()...)
}

func (m *v2UI) Setup(subscription partybus.Unsubscribable) error {
	// we still use the UI, but we want to suppress responding to events that would print out what is already
	// being logged.
	log.Set(discard.New())

	m.subscription = subscription
	m.program = tea.NewProgram(m, tea.WithOutput(os.Stderr), tea.WithInput(os.Stdin), tea.WithoutSignalHandler())
	m.running.Add(1)

	go func() {
		defer m.running.Done()
		if _, err := m.program.Run(); err != nil {
			log.Errorf("unable to start UI: %+v", err)
			bus.ExitWithInterrupt()
		}
	}()

	return nil
}

func (m *v2UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// note: we need a pointer receiver such that the same instance of UI used in Teardown is referenced (to keep finalize events)

	var cmds []tea.Cmd

	// allow for non-partybus UI updates (such as window size events). Note: these must not affect existing models,
	// that is the responsibility of the frame object on this UI object. The handler is a factory of models
	// which the frame is responsible for the lifecycle of. This update allows for injecting the initial state
	// of the world when creating those models.
	m.handler.OnMessage(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// today we treat esc and ctrl+c the same, but in the future when the worker has a graceful way to
		// cancel in-flight work via a context, we can wire up esc to this path with bus.Exit()
		case "esc", "ctrl+c":
			bus.ExitWithInterrupt()
			return m, tea.Quit
		}

	case partybus.Event:
		log.WithFields("component", "ui", "event", msg.Type).Trace("event")

		switch msg.Type {
		case event.Report, event.Notification:
			// keep these for when the UI is terminated to show to the screen (or perform other events)
			m.finalizeEvents = append(m.finalizeEvents, msg)

			// why not return tea.Quit here for exit events? because there may be UI components that still need the update-render loop.
			// for this reason we'll let the event loop call Teardown() which will explicitly wait for these components
			return m, nil
		}

		models, cmd := m.handler.Handle(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		for _, newModel := range models {
			if newModel == nil {
				continue
			}
			cmds = append(cmds, newModel.Init())
			f := m.frame.(frame.Frame)
			f.AppendModel(newModel)
			m.frame = f
		}
		// intentionally fallthrough to update the frame model
	}

	frameModel, cmd := m.frame.Update(msg)
	m.frame = frameModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m v2UI) View() string {
	return m.frame.View()
}

func (m *v2UI) Handle(e partybus.Event) error {
	if m.program != nil {
		m.program.Send(e)
	}
	return nil
}

func (m *v2UI) Teardown(force bool) error {
	defer func() {
		// allow for traditional logging to resume now that the UI is shutting down
		if logWrapper, ok := log.Get().(logger.Controller); ok {
			logWrapper.SetOutput(m.err)
		}
	}()

	if !force {
		m.handler.Wait()
		m.program.Quit()
		// typically in all cases we would want to wait for the UI to finish. However there are still error cases
		// that are not accounted for, resulting in hangs. For now, we'll just wait for the UI to finish in the
		// happy path only. There will always be an indication of the problem to the user via reporting the error
		// string from the worker (outside of the UI after teardown).
		m.running.Wait()
	} else {
		_ = runWithTimeout(250*time.Millisecond, func() error {
			m.handler.Wait()
			return nil
		})

		// it may be tempting to use Kill() however it has been found that this can cause the terminal to be left in
		// a bad state (where Ctrl+C and other control characters no longer works for future processes in that terminal).
		m.program.Quit()

		_ = runWithTimeout(250*time.Millisecond, func() error {
			m.running.Wait()
			return nil
		})
	}

	// TODO: allow for writing out the full log output to the screen (only a partial log is shown currently)
	// this needs coordination to know what the last frame event is to change the state accordingly (which isn't possible now)
	return writeEvents(m.out, m.err, m.quiet, m.format, m.finalizeEvents...)
}

func runWithTimeout(timeout time.Duration, fn func() error) (err error) {
	c := make(chan struct{}, 1)
	go func() {
		err = fn()
		c <- struct{}{}
	}()
	select {
	case <-c:
	case <-time.After(timeout):
		return fmt.Errorf("timed out after %v", timeout)
	}
	return err
}

func writeEvents(out, err io.Writer, quiet bool, format v2Format, events ...partybus.Event) error {
	handles := []struct {
		event        partybus.EventType
		respectQuiet bool
		writer       io.Writer
		dispatch     func(writer io.Writer, format v2Format, events ...partybus.Event) error
	}{
		{
			event:        event.Report,
			respectQuiet: false,
			writer:       out,
			dispatch:     writeReports,
		},
		{
			event:        event.Notification,
			respectQuiet: true,
			writer:       err,
			dispatch:     writeNotifications,
		},
	}

	var errs error
	for _, h := range handles {
		if quiet && h.respectQuiet {
			continue
		}

		for _, e := range events {
			if e.Type != h.event {
				continue
			}

			if err := h.dispatch(h.writer, format, e); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}
	return errs
}

func writeReports(writer io.Writer, format v2Format, events ...partybus.Event) error {
	var reports []string
	for _, e := range events {
		_, report, err := parser.ParseReport(e)
		if err != nil {
			log.WithFields("error", err).Warn("failed to gather final report")
			continue
		}

		// remove all whitespace padding from the end of the report
		reports = append(reports, strings.TrimRight(report, "\n ")+"\n")
	}

	// prevent the double new-line at the end of the report
	report := strings.Join(reports, "\n")

	if _, err := fmt.Fprint(writer, report); err != nil {
		return fmt.Errorf("failed to write final report to stdout: %w", err)
	}
	return nil
}

func writeNotifications(writer io.Writer, format v2Format, events ...partybus.Event) error {

	for _, e := range events {
		_, notification, err := parser.ParseNotification(e)
		if err != nil {
			log.WithFields("error", err).Warn("failed to parse notification")
			continue
		}

		if _, err := fmt.Fprintln(writer, format.Title.Render(notification)); err != nil {
			// don't let this be fatal
			log.WithFields("error", err).Warn("failed to write final notifications")
		}
	}
	return nil
}
