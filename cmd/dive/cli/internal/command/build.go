package command

import (
	"fmt"
	"github.com/anchore/clio"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/command/runtime"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
	"github.com/wagoodman/dive/dive"
)

type buildOptions struct {
	options.Application `yaml:",inline" mapstructure:",squash"`

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
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := setUI(app, opts.Application); err != nil {
				return fmt.Errorf("failed to set UI: %w", err)
			}
			return runtime.Run(
				cmd.Context(),
				runtime.Config{
					Source:     dive.ParseImageSource(opts.Analysis.ContainerEngine),
					BuildArgs:  args,
					Ci:         opts.CI.Enabled,
					CiRules:    opts.CI.Rules.List,
					ExportFile: opts.Export.JsonPath,
					UI:         opts.V1Preferences(),
				},
			)
		},
	}, opts)
}
