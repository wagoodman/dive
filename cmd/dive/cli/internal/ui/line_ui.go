package ui

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/wagoodman/go-partybus"

	"github.com/anchore/clio"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
	"github.com/wagoodman/dive/internal/bus/event"
	"github.com/wagoodman/dive/internal/bus/event/parser"
	"github.com/wagoodman/dive/internal/log"
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
	restoreLogs  func()
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
		n.restoreLogs = muteConsoleLogs()
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
		text, err := notificationText(e)
		if err != nil {
			log.WithFields("error", err, "event", fmt.Sprintf("%#v", e)).Warn("failed to parse event")
			return nil
		}

		// unlike the status lines this is nothing the logger says too, so it prints under -v as well
		fmt.Fprintln(n.err, n.format.Notification.Render(text))
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

func (n *LineUI) Teardown(_ bool) error {
	for _, e := range drain(n.subscription) {
		_ = n.Handle(e)
	}
	if n.restoreLogs != nil {
		n.restoreLogs()
		n.restoreLogs = nil
	}
	return nil
}

// notificationText is the text of a notification event, safe to print. Notifications carry image derived strings
// (paths, image config fields), so control characters other than newlines and tabs are escaped (as "\x1b") to keep
// them from driving the terminal.
func notificationText(e partybus.Event) (string, error) {
	_, text, err := parser.ParseNotification(e)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, r := range text {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			q := strconv.QuoteRune(r)
			b.WriteString(q[1 : len(q)-1])
			continue
		}
		b.WriteRune(r)
	}
	return b.String(), nil
}

// drain unsubscribes and returns the notifications and reports still queued for the UI. On an interrupt clio stops
// reading events without unsubscribing, which drops the ones already published (an extract error, the auto source
// notice). An explore event found here is released without starting the TUI.
func drain(sub partybus.Unsubscribable) []partybus.Event {
	s, ok := sub.(*partybus.Subscription)
	if !ok || s == nil {
		return nil
	}
	// nothing new is queued once unsubscribed, and the channel closes after what is already in it
	_ = s.Unsubscribe()

	var events []partybus.Event
	timeout := time.After(time.Second)
	for {
		select {
		case e, ok := <-s.Events():
			if !ok {
				return events
			}
			switch e.Type {
			case event.Notification, event.Report:
				events = append(events, e)
			case event.ExploreAnalysis:
				_, _, done, _ := parser.ParseExploreAnalysis(e)
				done()
			}
		case <-timeout:
			return events
		}
	}
}
