package ui

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/anchore/go-logger"
	"github.com/anchore/go-logger/adapter/discard"
	"github.com/wagoodman/go-partybus"

	v1 "github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/app"
	"github.com/wagoodman/dive/internal/bus/event/parser"
	"github.com/wagoodman/dive/internal/log"
)

// runExplore runs the gocui TUI for an ExploreAnalysis event and blocks until the user quits. This is the only place
// that reads the ExploreAnalysis payload, so a payload change only touches this function (and the parser).
func runExplore(e partybus.Event, prefs v1.Preferences) error {
	analysis, content, err := parser.ParseExploreAnalysis(e)
	if err != nil {
		return fmt.Errorf("unable to explore analysis: %w", err)
	}

	// ensure the logger will not interfere with the UI
	discardConsoleLogs()

	return app.Run(
		// TODO: this is not plumbed through from the command object...
		context.Background(),
		v1.Config{
			Content:     content,
			Analysis:    analysis,
			Preferences: prefs,
		},
	)
}

// discardConsoleLogs stops log lines from drawing over a UI. dive and stereoscope log through one logger, so muting
// its console output covers both. A logger that also writes to a file is left alone.
func discardConsoleLogs() {
	c, ok := log.Get().(logger.Controller)
	if !ok {
		log.Set(discard.New())
		return
	}
	if c.GetOutput() == os.Stderr {
		c.SetOutput(io.Discard)
	}
}
