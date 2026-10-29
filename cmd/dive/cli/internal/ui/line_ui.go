package ui

import (
	"fmt"
	"github.com/anchore/clio"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
	"github.com/wagoodman/dive/internal/bus/event"
	"github.com/wagoodman/dive/internal/bus/event/parser"
	"github.com/wagoodman/dive/internal/log"
	"github.com/wagoodman/go-partybus"
	"io"
	"os"
)

var _ clio.UI = (*LineUI)(nil)

type LineUI struct {
	cfg tui.Preferences
	out io.Writer
	err io.Writer

	subscription partybus.Unsubscribable
	quiet        bool
	verbosity    int
	format       format
}

type format struct {
	Title        lipgloss.Style
	Aux          lipgloss.Style
	Line         lipgloss.Style
	Notification lipgloss.Style
}

func NewLineUI(cfg tui.Preferences, out io.Writer, quiet bool, verbosity int) *LineUI {
	return &LineUI{
		cfg:       cfg,
		out:       out,
		err:       os.Stderr,
		quiet:     quiet,
		verbosity: verbosity,
		format: format{
			Title:        lipgloss.NewStyle().Bold(true).Width(30),
			Aux:          lipgloss.NewStyle().Faint(true),
			Notification: lipgloss.NewStyle().Foreground(lipgloss.Color("#A77BCA")),
		},
	}
}

func (n *LineUI) Setup(subscription partybus.Unsubscribable) error {
	if n.verbosity == 0 || n.quiet {
		// we still use the UI, but we want to suppress responding to events that would print out what is already
		// being logged.
		discardConsoleLogs()
	}

	// remove CI var from consideration when determining if we should use the UI
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(n.out, termenv.WithEnvironment(environWithoutCI{})))

	n.subscription = subscription
	return nil
}

func (n *LineUI) Handle(e partybus.Event) error {
	switch e.Type {
	case event.TaskStarted:
		if n.quiet {
			return nil
		}
		prog, task, err := parser.ParseTaskStarted(e)
		if err != nil {
			log.WithFields("error", err, "event", fmt.Sprintf("%#v", e)).Warn("failed to parse event")
		}

		var aux string
		stage := prog.Stage()
		switch {
		case task.Context != "":
			aux = task.Context
		case stage != "":
			aux = stage
		}

		if aux != "" {
			aux = n.format.Aux.Render(aux)
		}

		n.writeToStderr(n.format.Title.Render(task.Title.Default) + aux)
	case event.Notification:
		if n.quiet {
			return nil
		}
		_, text, err := parser.ParseNotification(e)
		if err != nil {
			log.WithFields("error", err, "event", fmt.Sprintf("%#v", e)).Warn("failed to parse event")
		}

		n.writeToStderr(n.format.Notification.Render(text))
	case event.Report:
		if n.quiet {
			return nil
		}
		_, text, err := parser.ParseReport(e)
		if err != nil {
			log.WithFields("error", err, "event", fmt.Sprintf("%#v", e)).Warn("failed to parse event")
		}

		n.writeToStderr("")
		n.writeToStdout(text)
	case event.ExploreAnalysis:
		return runExplore(e, n.cfg)
	}
	return nil
}

func (n *LineUI) writeToStdout(s string) {
	fmt.Fprintln(n.out, s)
}

func (n *LineUI) writeToStderr(s string) {
	if n.quiet || n.verbosity > 0 {
		// we've been told to not report anything or that we're in verbose mode thus the logger should report all info.
		// This only applies to status like info on stderr, not to primary reports on stdout.
		return
	}
	fmt.Fprintln(n.err, s)
}

func (n LineUI) Teardown(_ bool) error {
	return nil
}
