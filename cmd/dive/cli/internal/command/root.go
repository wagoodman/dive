package command

import (
	"fmt"
	"github.com/anchore/clio"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
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
			if opts.Development.UseStereoscope {
				return v2FetchImage(cmd.Context(), opts.Application, app)
			}
			return v1FetchImage(cmd.Context(), opts.Application, app)
		},
	}, opts)
}
