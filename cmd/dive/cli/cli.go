package cli

import (
	"github.com/anchore/clio"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/command"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui"
	"github.com/wagoodman/dive/internal/bus"
	"github.com/wagoodman/dive/internal/log"
	"io"
	"os"
)

func Application(id clio.Identification) clio.Application {
	app, _ := create(id, os.Stdout)
	return app
}

func Command(id clio.Identification) *cobra.Command {
	_, cmd := create(id, os.Stdout)
	return cmd
}

func create(id clio.Identification, out io.Writer) (clio.Application, *cobra.Command) {
	clioCfg := clio.NewSetupConfig(id).
		WithGlobalConfigFlag().   // add persistent -c <path> for reading an application config from
		WithGlobalLoggingFlags(). // add persistent -v and -q flags tied to the logging config
		WithConfigInRootHelp().   // --help on the root command renders the full application config in the help text
		//WithUIConstructor(
		//	// select a UI based on the logging configuration and state of stdin (if stdin is a tty)
		//	func(cfg clio.Config) (*clio.UICollection, error) {
		//		noUI := ui.None(out, cfg.Log.Quiet)
		//		if !cfg.Log.AllowUI(os.Stdin) || cfg.Log.Quiet {
		//			return clio.NewUICollection(noUI), nil
		//		}
		//
		//		return clio.NewUICollection(
		//			ui.NewV1UI(out, cfg.Log.Quiet),
		//			noUI, // fallback incase the v1 UI fails
		//		), nil
		//	},
		//).
		WithUI(ui.None()).
		WithInitializers(
			func(state *clio.State) error {
				bus.Set(state.Bus)
				log.Set(state.Logger)

				//stereoscope.SetBus(state.Bus)
				//stereoscope.SetLogger(state.Logger)
				return nil
			},
		)
	//WithPostRuns(func(_ *clio.State, _ error) {
	//	stereoscope.Cleanup()
	//})

	app := clio.New(*clioCfg)

	rootCmd := command.Root(app)

	rootCmd.AddCommand(
		clio.VersionCommand(id),
		clio.ConfigCommand(app, nil),
		command.Build(app),
	)

	return app, rootCmd
}
