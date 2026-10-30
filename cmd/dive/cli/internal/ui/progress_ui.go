// copied and adapted from syft v1.55.0 (github.com/anchore/syft, cmd/syft/internal/ui and cmd/syft/internal/clio_setup_config.go)

package ui

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
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
// the plain line based UI (no TTY, piped stdin, --quiet, -v, NO_TTY or CI set, or plain for output meant for scripts,
// like --ci and --json). Either one hands the terminal to the gocui TUI on an ExploreAnalysis event.
func New(prefs tui.Preferences, out io.Writer, quiet bool, verbosity int, plain bool) clio.UI {
	lc := clio.LoggingConfig{Quiet: quiet, Verbosity: verbosity}
	if plain || quiet || isCI() || !lc.AllowUI(os.Stdin) {
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
	prefs        tui.Preferences
	out          io.Writer
	err          io.Writer
	in           io.Reader
	program      *tea.Program
	running      *sync.WaitGroup
	exited       chan struct{}
	stopped      bool
	subscription partybus.Unsubscribable
	restoreLogs  func()
	// interrupted is set by ctrl-c in the bubbletea loop and read by the event loop
	interrupted atomic.Bool

	// finalizeEvents are appended in the bubbletea loop and read on teardown
	eventsLock     sync.Mutex
	finalizeEvents []partybus.Event

	handler *bubbly.HandlerCollection
	frame   tea.Model
}

func newProgressUI(prefs tui.Preferences, out io.Writer, handlers ...bubbly.EventHandler) *progressUI {
	return &progressUI{
		prefs:   prefs,
		out:     out,
		err:     os.Stderr,
		in:      os.Stdin,
		handler: bubbly.NewHandlerCollection(handlers...),
		frame:   frame.New(),
		running: &sync.WaitGroup{},
		exited:  make(chan struct{}),
	}
}

func (m *progressUI) Setup(subscription partybus.Unsubscribable) error {
	// the progress UI owns stderr while it runs
	m.restoreLogs = muteConsoleLogs()

	// remove CI var from consideration when determining colors
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(m.err, termenv.WithEnvironment(environWithoutCI{})))

	m.subscription = subscription
	m.program = tea.NewProgram(m, tea.WithOutput(m.err), tea.WithInput(m.in), tea.WithoutSignalHandler())
	m.running.Add(1)

	go func() {
		defer m.running.Done()
		defer close(m.exited)
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
		if m.interrupted.Load() {
			// ctrl-c asked to exit, the exit event is only queued behind this one
			_, _, done, _ := parser.ParseExploreAnalysis(e)
			done()
			return nil
		}
		return runExplore(e, m.prefs)
	}
	if !m.stopped && e.Type == event.Notification && m.printAbove(e) {
		return nil
	}
	if m.stopped || e.Type == event.Notification {
		// the program is gone (handed off to the TUI) or going, keep what teardown prints
		if e.Type == event.Report || e.Type == event.Notification {
			m.keep(e)
		}
		return nil
	}
	if m.program != nil {
		m.program.Send(e)
	}
	return nil
}

func (m *progressUI) keep(events ...partybus.Event) {
	m.eventsLock.Lock()
	m.finalizeEvents = append(m.finalizeEvents, events...)
	m.eventsLock.Unlock()
}

// printAbove prints a notification above the progress rows, so it shows while the image loads (the auto source
// notice) instead of after the TUI. It reports whether bubbletea took the line, which it can't once it quit.
func (m *progressUI) printAbove(e partybus.Event) bool {
	text, err := notificationText(e)
	if m.program == nil || err != nil || m.interrupted.Load() {
		return false
	}
	// Send gives up silently once the program is done, so follow the line with a flush: bubbletea handles messages
	// in order, a handled flush means the line was queued for printing too
	m.program.Send(tea.Println(styleNotification(text))())
	c := make(flushMsg)
	m.program.Send(c)
	select {
	case <-c:
		return true
	case <-m.exited:
		select {
		case <-c:
			return true
		default:
			return false
		}
	}
}

// stop renders the last frame and quits the program, waiting at most the given time for task rows to finish. It
// blocks until bubbletea has restored the terminal.
func (m *progressUI) stop(wait time.Duration) {
	if m.stopped || m.program == nil {
		return
	}
	m.stopped = true

	m.flush(wait)
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
		m.flush(time.Second)
		m.handler.Wait()
		m.program.Quit()
		m.running.Wait()
	default:
		m.flush(250 * time.Millisecond)
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

	if m.restoreLogs != nil {
		m.restoreLogs()
		m.restoreLogs = nil
	}

	// clio may stop reading on an interrupt with events still queued
	m.keep(drain(m.subscription)...)

	m.eventsLock.Lock()
	events := m.finalizeEvents
	m.eventsLock.Unlock()
	return writeEvents(m.out, m.err, events...)
}

// flushMsg is closed once bubbletea handles it, which is after every event sent before it
type flushMsg chan struct{}

// flush waits (at most timeout) until bubbletea has handled every event already sent to it, so the task rows those
// events start are counted by the handler's wait group before anyone waits on it.
func (m *progressUI) flush(timeout time.Duration) {
	c := make(flushMsg)
	go m.program.Send(c)
	select {
	case <-c:
	case <-m.exited:
	case <-time.After(timeout):
	}
}

// bubbletea.Model functions

func (m *progressUI) Init() tea.Cmd {
	return m.frame.Init()
}

func (m *progressUI) RespondsTo() []partybus.EventType {
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
	case flushMsg:
		close(msg)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		// today we treat esc and ctrl+c the same, but in the future when the worker has a graceful way to
		// cancel in-flight work via a context, we can wire up esc to this path with bus.Exit()
		case "esc", "ctrl+c":
			m.interrupted.Store(true)
			bus.ExitWithInterrupt()
			return m, tea.Quit
		}

	case partybus.Event:
		log.WithFields("component", "ui", "event", msg.Type).Trace("event")

		switch msg.Type {
		case event.Report, event.Notification:
			// keep these for when the UI is terminated to show to the screen (or perform other events)
			m.keep(msg)

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

func (m *progressUI) View() string {
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
	for _, e := range events {
		switch e.Type {
		case event.Notification:
			text, err := notificationText(e)
			if err != nil {
				log.WithFields("error", err).Warn("failed to parse notification")
				continue
			}
			fmt.Fprintln(errOut, styleNotification(text))
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

func styleNotification(text string) string {
	// 13 = high intensity magenta (ANSI 16 bit code)
	return lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Render(text)
}
