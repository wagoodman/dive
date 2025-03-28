package command

import (
	"github.com/wagoodman/dive/cmd/dive/cli/options"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/runtime"

	"github.com/spf13/cobra"

	"github.com/anchore/clio"
)

type buildOptions struct {
	options.Application

	// reserved for future use of build-only flags
}

func Build(app clio.Application) *cobra.Command {
	opts := &buildOptions{
		Application: options.DefaultApplication(),
	}
	return app.SetupCommand(&cobra.Command{
		Use:                "build [any valid `docker build` arguments]",
		Short:              "Builds and analyzes a docker image from a Dockerfile (this is a thin wrapper for the `docker build` command).",
		DisableFlagParsing: true,
		RunE: func(_ *cobra.Command, args []string) error {
			return runBuild(opts, args)
		},
	}, opts)
}

func runBuild(opts *buildOptions, args []string) error {
	runtime.Run(
		runtime.Options{
			Source:       dive.ParseImageSource(opts.Analysis.ContainerEngine),
			BuildArgs:    args,
			Ci:           opts.CI.Enabled,
			CiRules:      opts.CI.Rules.List,
			IgnoreErrors: opts.Analysis.IgnoreErrors,
			ExportFile:   opts.Export.JsonPath,
			KeyBindings:  opts.UI.Keybinding.Bindings,
		},
	)

	return nil
}
