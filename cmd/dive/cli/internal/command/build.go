package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/anchore/clio"
	"github.com/spf13/cobra"

	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
	"github.com/wagoodman/dive/internal/bus"
	"github.com/wagoodman/dive/internal/bus/event/payload"
	"github.com/wagoodman/dive/internal/log"
	"github.com/wagoodman/dive/internal/source"
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
		RunE: escaped(func(cmd *cobra.Command, args []string) error {
			// the build subprocess writes to the terminal, so no UI may take it over
			if err := setUI(app, opts.Application, lineUI); err != nil {
				return fmt.Errorf("failed to set UI: %w", err)
			}

			ref, err := build(cmd.Context(), opts.Analysis.Source, args)
			if err != nil {
				return fmt.Errorf("cannot build image: %w", err)
			}

			// the ref has a scheme, so the configured source doesn't apply to it
			return run(cmd.Context(), opts.Application, ref)
		}),
	}, opts)
}

func build(ctx context.Context, engine string, args []string) (string, error) {
	// only docker and podman can build (source.Build refuses the rest), auto builds with docker
	if engine, _ = source.Normalize(engine); engine == source.Auto {
		engine = source.Docker
	}

	log.Info("building image")
	log.Debugf("└── %s", strings.Join(args, " "))

	mon := bus.StartTask(payload.GenericTask{
		Title: payload.Title{
			Default:      "Building image",
			WhileRunning: "Building image",
			OnSuccess:    "Built image",
			OnFail:       "Failed to build image",
		},
		Context: "... " + strings.Join(args, " "),
	})

	ref, err := source.Build(ctx, engine, args)
	finish(mon, err)
	return ref, err
}
