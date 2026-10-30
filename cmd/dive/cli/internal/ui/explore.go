package ui

import (
	"fmt"
	"io"

	"github.com/anchore/go-logger"
	"github.com/wagoodman/go-partybus"

	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui/app"
	"github.com/wagoodman/dive/internal/bus/event/parser"
	"github.com/wagoodman/dive/internal/log"
)

// runExplore runs the gocui TUI for an ExploreAnalysis event and blocks until the user quits or the command's context
// is done. This is the only place that reads the ExploreAnalysis payload, so a payload change only touches this
// function (and the parser).
func runExplore(e partybus.Event, prefs tui.Preferences) error {
	analysis, ctx, done, err := parser.ParseExploreAnalysis(e)
	// the publisher waits on this before cleaning up the image
	defer done()
	if err != nil {
		return fmt.Errorf("unable to explore analysis: %w", err)
	}
	if ctx.Err() != nil {
		// interrupted before the TUI got the terminal
		return nil
	}

	// ensure the logger will not interfere with the UI
	restore := muteConsoleLogs()
	defer restore()

	return app.Run(ctx, tui.Config{
		Analysis:    analysis,
		Preferences: prefs,
	})
}

// muteConsoleLogs stops log lines from drawing over a UI and returns the func that brings them back. dive and
// stereoscope log through clio's logger, which also reports the command's final error once the UI is gone, so the
// restore has to run when the UI hands the terminal back. A configured log file shares the writer with stderr, so it
// misses the lines in between too.
func muteConsoleLogs() (restore func()) {
	c, ok := log.Get().(logger.Controller)
	if !ok {
		// clio's logrus logger and the discard default both are controllers, there is nothing else to mute
		return func() {}
	}
	out := c.GetOutput()
	c.SetOutput(io.Discard)
	return func() { c.SetOutput(out) }
}
