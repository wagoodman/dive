package command

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/anchore/clio"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

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
	cmd := app.SetupCommand(&cobra.Command{
		Use:   "build [dive flags --] [any valid `docker build` arguments]",
		Short: "Builds and analyzes a docker image from a Dockerfile (this is a thin wrapper for the `docker build` command).",
		Long: `Builds and analyzes a docker image from a Dockerfile (this is a thin wrapper for the ` + "`docker build`" + ` command).

All arguments go to the build, unless there is a "--": then the arguments before it are dive flags and the ones
after it go to the build, e.g. "dive build --ci -- -t app .".`,
		DisableFlagParsing: true,
		RunE: escaped(func(cmd *cobra.Command, args []string) error {
			// the build subprocess writes to the terminal, so no UI may take it over
			if err := setUI(app, opts.Application, lineUI); err != nil {
				return fmt.Errorf("failed to set UI: %w", err)
			}

			_, buildArgs := splitBuildArgs(args)
			ref, err := build(cmd.Context(), opts.Analysis.Source, buildArgs)
			if err != nil {
				return fmt.Errorf("cannot build image: %w", err)
			}

			// the ref has a scheme, so the configured source doesn't apply to it
			return run(cmd.Context(), opts.Application, ref)
		}),
	}, opts)

	// flag parsing is off so the build args pass through untouched. The dive flags are parsed here instead, ahead
	// of the config load in the setup PreRunE
	setup := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if slices.Contains(args, "-h") || slices.Contains(args, "--help") {
			// cobra shows the help for this
			return pflag.ErrHelp
		}
		diveArgs, _ := splitBuildArgs(args)
		cmd.InheritedFlags() // merges the root's persistent flags (--config, -v, ...) into cmd.Flags()
		if err := cmd.Flags().Parse(diveArgs); err != nil {
			return err
		}
		if cmd.Flags().NArg() > 0 {
			return fmt.Errorf("unexpected arguments before \"--\": %s", strings.Join(cmd.Flags().Args(), " "))
		}
		return setup(cmd, args)
	}
	return cmd
}

// splitBuildArgs splits the args at the first "--" into dive flags and build args. Without one, everything is a
// build arg.
func splitBuildArgs(args []string) (diveArgs, buildArgs []string) {
	if i := slices.Index(args, "--"); i >= 0 {
		return args[:i], args[i+1:]
	}
	return nil, args
}

// buildEngine picks the engine to build with for the configured source. Only docker and podman can build:
// containers-storage is podman's store, every other source (auto, archives, registry, ...) builds with docker.
func buildEngine(src string) string {
	if src, _ = source.Normalize(src); src == source.Podman || src == source.ContainersStorage {
		return source.Podman
	}
	return source.Docker
}

func build(ctx context.Context, src string, args []string) (string, error) {
	engine := buildEngine(src)

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
