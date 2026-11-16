package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/anchore/clio"
	"github.com/anchore/go-logger"
	"github.com/anchore/go-logger/adapter/redact"
	"github.com/anchore/stereoscope"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/command"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui"
	"github.com/wagoodman/dive/internal/bus"
	"github.com/wagoodman/dive/internal/log"
)

func Application(id clio.Identification) clio.Application {
	app, _ := create(id)
	return app
}

func Command(id clio.Identification) *cobra.Command {
	_, cmd := create(id)
	return cmd
}

func create(id clio.Identification) (clio.Application, *cobra.Command) {
	clioCfg := clio.NewSetupConfig(id).
		WithGlobalConfigFlag().   // add persistent -c <path> for reading an application config from
		WithGlobalLoggingFlags(). // add persistent -v and -q flags tied to the logging config
		WithConfigInRootHelp().   // --help on the root command renders the full application config in the help text
		WithUI(ui.None()).
		WithLoggerConstructor(newLogger).
		WithInitializers(
			func(state *clio.State) error {
				bus.Set(state.Bus)
				log.Set(state.Logger)

				// stereoscope publishes pull/fetch/read events on the same bus the UI listens to
				stereoscope.SetBus(state.Bus)
				stereoscope.SetLogger(state.Logger.Nested("from-lib", "stereoscope"))
				return nil
			},
		)

	app := clio.New(*clioCfg)

	rootCmd := command.Root(app)

	rootCmd.AddCommand(
		clio.VersionCommand(id),
		clio.ConfigCommand(app, nil),
		command.Build(app),
	)

	return app, rootCmd
}

// newLogger is clio's default logger, except that it opens the log file itself and records it with log.SetFile. clio
// tees the console and the file through one writer, so muting the console while a UI draws would drop the file too.
func newLogger(cfg clio.Config, store redact.Store) (logger.Logger, error) {
	if cfg.Log == nil || cfg.Log.FileLocation == "" {
		return clio.DefaultLogger(cfg, store)
	}
	path := cfg.Log.FileLocation
	consoleOnly := *cfg.Log
	consoleOnly.FileLocation = ""
	cfg.Log = &consoleOnly

	l, err := clio.DefaultLogger(cfg, store)
	if err != nil {
		return nil, err
	}
	c, ok := l.(logger.Controller)
	if !ok {
		return nil, fmt.Errorf("unable to setup log file: logger %T has no settable output", l)
	}
	// same flags and mode clio opens it with
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, fmt.Errorf("unable to setup log file: %w", err)
	}
	out := io.Writer(f)
	if !consoleOnly.Quiet {
		out = io.MultiWriter(os.Stderr, f)
	}
	c.SetOutput(out)
	log.SetFile(f)
	return l, nil
}
