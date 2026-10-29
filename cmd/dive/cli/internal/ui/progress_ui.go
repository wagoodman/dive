// copied and adapted from syft v1.55.0 (github.com/anchore/syft, cmd/syft/internal/ui and cmd/syft/internal/clio_setup_config.go)

package ui

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/wagoodman/go-partybus"

	"github.com/anchore/bubbly"
	"github.com/anchore/bubbly/bubbles/frame"
	"github.com/anchore/clio"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/handler"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
	"github.com/wagoodman/dive/internal/bus"
	"github.com/wagoodman/dive/internal/bus/event"
	"github.com/wagoodman/dive/internal/bus/event/parser"
	"github.com/wagoodman/dive/internal/log"
)

// New returns the UI for a command run: the bubbletea progress UI when stderr is an interactive terminal, otherwise
// the plain line based UI (no TTY, piped stdin, --quiet, -v, NO_TTY or CI set). Either one hands the terminal to the
// gocui TUI on an ExploreAnalysis event.
func New(prefs tui.Preferences, out io.Writer, quiet bool, verbosity int) clio.UI {
	lc := clio.LoggingConfig{Quiet: quiet, Verbosity: verbosity}
	if quiet || isCI() || !lc.AllowUI(os.Stdin) {
		return NewLineUI(prefs, out, quiet, verbosity)
	}
	return newProgressUI(prefs, out, handler.New(handler.DefaultHandlerConfig()))
}

func isCI() bool {
	switch os.Getenv("CI") {
	case "true", "1", "yes":
		return true
	}
	return false
}

var _ interface {
	tea.Model
	partybus.Responder
	clio.UI
} = (*progressUI)(nil)

type progressUI struct {
	prefs          tui.Preferences
	out            io.Writer
	err            io.Writer
	program        *tea.Program
	running        *sync.WaitGroup
	stopped        bool
	subscription   partybus.Unsubscribable
	finalizeEvents []partybus.Event

	handler *bubbly.HandlerCollection
	frame   tea.Model
}

func newProgressUI(prefs tui.Preferences, out io.Writer, handlers ...bubbly.EventHandler) *progressUI {
	return &progressUI{
		prefs:   prefs,
		out:     out,
		err:     os.Stderr,
		handler: bubbly.NewHandlerCollection(handlers...),
		frame:   frame.New(),
		running: &sync.WaitGroup{},
	}
}

func (m *progressUI) Setup(subscription partybus.Unsubscribable) error {
	// the progress UI owns stderr while it runs
	discardConsoleLogs()

	// remove CI var from consideration when determining colors
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(m.err, termenv.WithEnvironment(environWithoutCI{})))

	m.subscription = subscription
	m.program = tea.NewProgram(m, tea.WithOutput(m.err), tea.WithInput(os.Stdin), tea.WithoutSignalHandler())
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

func (m *progressUI) Handle(e partybus.Event) error {
	if e.Type == event.ExploreAnalysis {
		// the progress program has to be gone (terminal restored, stdin released) before gocui takes over
		m.stop(time.Second)
		return runExplore(e, m.prefs)
	}
	if m.program != nil {
		m.program.Send(e)
	}
	return nil
}

// stop renders the last frame and quits the program, waiting at most the given time for task rows to finish. It
// blocks until bubbletea has restored the terminal.
func (m *progressUI) stop(wait time.Duration) {
	if m.stopped || m.program == nil {
		return
	}
	m.stopped = true

	_ = runWithTimeout(wait, func() error {
		m.handler.Wait()
		return nil
	})

	// it may be tempting to use Kill() however it has been found that this can cause the terminal to be left in
	// a bad state (where Ctrl+C and other control characters no longer work for future processes in that terminal).
	m.program.Quit()
	m.running.Wait()
}

func (m *progressUI) Teardown(force bool) error {
	switch {
	case m.stopped || m.program == nil:
		// already handed off to the TUI (or never started)
	case !force:
		// typically in all cases we would want to wait for the UI to finish. However there are still error cases
		// that are not accounted for, resulting in hangs. For now, we'll just wait for the UI to finish in the
		// happy path only. There will always be an indication of the problem to the user via reporting the error
		// string from the worker (outside of the UI after teardown).
		m.handler.Wait()
		m.program.Quit()
		m.running.Wait()
	default:
		_ = runWithTimeout(250*time.Millisecond, func() error {
			m.handler.Wait()
			return nil
		})
		m.program.Quit()
		_ = runWithTimeout(250*time.Millisecond, func() error {
			m.running.Wait()
			return nil
		})
	}
	m.stopped = true

	return writeEvents(m.out, m.err, m.finalizeEvents...)
}

// bubbletea.Model functions

func (m progressUI) Init() tea.Cmd {
	return m.frame.Init()
}

func (m progressUI) RespondsTo() []partybus.EventType {
	return append([]partybus.EventType{
		event.Report,
		event.Notification,
		event.ExploreAnalysis,
	}, m.handler.RespondsTo()...)
}

func (m *progressUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.frame.(*frame.Frame).AppendModel(newModel)
		}
		// intentionally fallthrough to update the frame model
	}

	frameModel, cmd := m.frame.Update(msg)
	m.frame = frameModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m progressUI) View() string {
	return m.frame.View()
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

// writeEvents prints the events held back while the progress UI owned the terminal
func writeEvents(out, errOut io.Writer, events ...partybus.Event) error {
	// 13 = high intensity magenta (ANSI 16 bit code)
	notificationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("13"))

	for _, e := range events {
		switch e.Type {
		case event.Notification:
			_, text, err := parser.ParseNotification(e)
			if err != nil {
				log.WithFields("error", err).Warn("failed to parse notification")
				continue
			}
			fmt.Fprintln(errOut, notificationStyle.Render(text))
		case event.Report:
			_, text, err := parser.ParseReport(e)
			if err != nil {
				log.WithFields("error", err).Warn("failed to gather final report")
				continue
			}
			// same text as the line based UI prints, so --ci output doesn't depend on the UI in use
			if _, err := fmt.Fprintln(out, text); err != nil {
				return fmt.Errorf("failed to write final report to stdout: %w", err)
			}
		}
	}
	return nil
}
