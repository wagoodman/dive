package ui

import (
	"context"
	"fmt"
	"github.com/anchore/clio"
	"github.com/anchore/go-logger/adapter/discard"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	v1 "github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/app"
	"github.com/wagoodman/dive/internal/bus/event"
	"github.com/wagoodman/dive/internal/bus/event/parser"
	"github.com/wagoodman/dive/internal/log"
	"github.com/wagoodman/go-partybus"
	"io"
	"os"
	"strings"
)

var _ clio.UI = (*V1UI)(nil)

type V1UI struct {
	cfg v1.Preferences
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

func NewV1UI(cfg v1.Preferences, out io.Writer, quiet bool, verbosity int) *V1UI {
	return &V1UI{
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

func (n *V1UI) Setup(subscription partybus.Unsubscribable) error {
	if n.verbosity == 0 || n.quiet {
		// we still use the UI, but we want to suppress responding to events that would print out what is already
		// being logged.
		log.Set(discard.New())
	}

	// remove CI var from consideration when determining if we should use the UI
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(n.out, termenv.WithEnvironment(environWithoutCI{})))

	n.subscription = subscription
	return nil
}

var _ termenv.Environ = (*environWithoutCI)(nil)

type environWithoutCI struct {
}

func (e environWithoutCI) Environ() []string {
	var out []string
	for _, s := range os.Environ() {
		if strings.HasPrefix(s, "CI=") {
			continue
		}
		out = append(out, s)
	}
	return out
}

func (e environWithoutCI) Getenv(s string) string {
	if s == "CI" {
		return ""
	}
	return os.Getenv(s)
}

func (n *V1UI) Handle(e partybus.Event) error {
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
		analysis, content, err := parser.ParseExploreAnalysis(e)
		if err != nil {
			log.WithFields("error", err, "event", fmt.Sprintf("%#v", e)).Warn("failed to parse event")
		}

		// ensure the logger will not interfere with the UI
		log.Set(discard.New())

		return app.Run(
			// TODO: this is not plumbed through from the command object...
			context.Background(),
			v1.Config{
				Content:     content,
				Analysis:    analysis,
				Preferences: n.cfg,
			},
		)
	}
	return nil
}

func (n *V1UI) writeToStdout(s string) {
	fmt.Fprintln(n.out, s)
}

func (n *V1UI) writeToStderr(s string) {
	if n.quiet || n.verbosity > 0 {
		// we've been told to not report anything or that we're in verbose mode thus the logger should report all info.
		// This only applies to status like info on stderr, not to primary reports on stdout.
		return
	}
	fmt.Fprintln(n.err, s)
}

func (n V1UI) Teardown(_ bool) error {
	return nil
}
