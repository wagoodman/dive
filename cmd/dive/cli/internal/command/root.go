package command

import (
	"fmt"
	"github.com/anchore/clio"
	"github.com/anchore/go-logger/adapter/discard"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/command/runtime"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui"
	"github.com/wagoodman/dive/internal/log"
	"os"
)

type rootOptions struct {
	options.Application `yaml:",inline" mapstructure:",squash"`

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
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := setUI(app, opts.Application); err != nil {
				return fmt.Errorf("failed to set UI: %w", err)
			}
			return runtime.Run(
				cmd.Context(),
				runtime.Config{
					Image:      opts.Analysis.Image,
					Source:     opts.Analysis.Source,
					Ci:         opts.CI.Enabled,
					CiRules:    opts.CI.Rules.List,
					ExportFile: opts.Export.JsonPath,
					UI:         opts.V1Preferences(),
				},
			)
		},
	}, opts)
}

func setUI(app clio.Application, opts options.Application) error {
	log.Set(discard.New())

	type Stater interface {
		State() *clio.State
	}

	state := app.(Stater).State()

	ux := ui.NewV1UI(opts.V1Preferences(), os.Stdout, state.Config.Log.Quiet)
	return state.UI.Replace(ux)
}
