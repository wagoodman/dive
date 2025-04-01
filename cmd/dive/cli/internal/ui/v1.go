package ui

import (
	"context"
	"fmt"
	"github.com/anchore/clio"
	"github.com/charmbracelet/lipgloss"
	v1 "github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/app"
	"github.com/wagoodman/dive/internal/bus/event"
	"github.com/wagoodman/dive/internal/bus/event/parser"
	"github.com/wagoodman/dive/internal/log"
	"github.com/wagoodman/go-partybus"
	"io"
	"os"
)

var _ clio.UI = (*V1UI)(nil)

type V1UI struct {
	cfg v1.Preferences
	out io.Writer
	err io.Writer

	subscription partybus.Unsubscribable
	quiet        bool
	format       format
}

type format struct {
	Title        lipgloss.Style
	Aux          lipgloss.Style
	Line         lipgloss.Style
	Notification lipgloss.Style
}

func NewV1UI(cfg v1.Preferences, out io.Writer, quiet bool) *V1UI {
	return &V1UI{
		cfg:   cfg,
		out:   out,
		err:   os.Stderr,
		quiet: quiet,
		format: format{
			Title:        lipgloss.NewStyle().Bold(true).Width(30),
			Aux:          lipgloss.NewStyle().Faint(true),
			Notification: lipgloss.NewStyle().Foreground(lipgloss.Color("#A77BCA")),
		},
	}
}

func (n *V1UI) Setup(subscription partybus.Unsubscribable) error {
	n.subscription = subscription
	return nil
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

		n.writeToStdout(text)
	case event.ExploreAnalysis:
		analysis, content, err := parser.ParseExploreAnalysis(e)
		if err != nil {
			log.WithFields("error", err, "event", fmt.Sprintf("%#v", e)).Warn("failed to parse event")
		}
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
	fmt.Fprintln(n.err, s)
}

func (n V1UI) Teardown(_ bool) error {
	return nil
}
