package command

import (
	"fmt"
	"github.com/anchore/clio"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/cmd/dive/cli/options"
	"github.com/wagoodman/dive/runtime"
)

type rootOptions struct {
	options.Application

	// reserved for future use of root-only flags
}

func Root(app clio.Application) *cobra.Command {
	opts := &rootOptions{
		Application: options.DefaultApplication(),
	}
	return app.SetupRootCommand(&cobra.Command{
		Use:   "dive [IMAGE]",
		Short: "Docker Image Visualizer & Explorer",
		Long: `This tool provides a way to discover and explore the contents of a docker image. Additionally the tool estimates
the amount of wasted space and identifies the offending files from the image.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("exactly one argument is required")
			}
			opts.Analysis.Image = args[0]
			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runRoot(opts)
		},
	}, opts)
}

func runRoot(opts *rootOptions) error {

	runtime.Run(
		runtime.Options{
			Image:        opts.Analysis.Image,
			Source:       opts.Analysis.Source,
			Ci:           opts.CI.Enabled,
			CiRules:      opts.CI.Rules.List,
			IgnoreErrors: opts.Analysis.IgnoreErrors,
			ExportFile:   opts.Export.JsonPath,
			KeyBindings:  opts.UI.Keybinding.Bindings,
		},
	)

	return nil
}
