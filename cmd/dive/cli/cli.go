package cli

import (
	"github.com/anchore/clio"
	"github.com/spf13/cobra"
	command2 "github.com/wagoodman/dive/cmd/dive/cli/internal/command"
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
		WithConfigInRootHelp()    // --help on the root command renders the full application config in the help text
	//WithUIConstructor(
	//	// select a UI based on the logging configuration and state of stdin (if stdin is a tty)
	//	func(cfg clio.Config) (*clio.UICollection, error) {
	//		noUI := ui.None(out, cfg.Log.Quiet)
	//		if !cfg.Log.AllowUI(os.Stdin) || cfg.Log.Quiet {
	//			return clio.NewUICollection(noUI), nil
	//		}
	//
	//		return clio.NewUICollection(
	//			ui.New(out, cfg.Log.Quiet,
	//				handler.New(handler.DefaultHandlerConfig()),
	//			),
	//			noUI,
	//		), nil
	//	},
	//).
	//WithInitializers(
	//	func(state *clio.State) error {
	//		// clio is setting up and providing the bus, redact store, and logger to the application. Once loaded,
	//		// we can hoist them into the internal packages for global use.
	//		//stereoscope.SetBus(state.Bus)
	//		//bus.Set(state.Bus)
	//		//
	//		//redact.Set(state.RedactStore)
	//		//
	//		//log.Set(state.Logger)
	//		//stereoscope.SetLogger(state.Logger)
	//		return nil
	//	},
	//)
	//WithPostRuns(func(_ *clio.State, _ error) {
	//	stereoscope.Cleanup()
	//})

	app := clio.New(*clioCfg)

	rootCmd := command2.Root(app)

	// add sub-commands
	rootCmd.AddCommand(
		// basic commands
		clio.VersionCommand(id),
		clio.ConfigCommand(app, nil),

		// application commands
		command2.Build(app),
	)

	return app, rootCmd
}
