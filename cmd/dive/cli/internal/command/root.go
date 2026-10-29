package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/anchore/clio"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/command/adapter"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/dive/source"
	"github.com/wagoodman/dive/internal/bus"
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

			resolver, err := dive.GetImageResolver(opts.Analysis.Source)
			if err != nil {
				return fmt.Errorf("cannot determine image provider to fetch from: %w", err)
			}

			ctx := cmd.Context()

			img, err := adapter.ImageResolver(resolver).Fetch(ctx, opts.Analysis.Image)
			if err != nil {
				return fmt.Errorf("cannot load image: %w", err)
			}

			return run(ctx, opts.Application, img, opts.Analysis.Reference)
		},
	}, opts)
}

func setUI(app clio.Application, opts options.Application) error {
	type Stater interface {
		State() *clio.State
	}

	state := app.(Stater).State()

	ux := ui.NewV1UI(opts.V1Preferences(), os.Stdout, state.Config.Log.Quiet, state.Config.Log.Verbosity)
	return state.UI.Replace(ux)
}

// ref is what the TUI bridge loads through dive/source (the raw reference, or the ID of a built image)
func run(ctx context.Context, opts options.Application, img *image.Image, ref string) error {
	analysis, err := adapter.NewAnalyzer().Analyze(ctx, img)
	if err != nil {
		return fmt.Errorf("cannot analyze image: %w", err)
	}

	if opts.Export.JsonPath != "" {
		if err := adapter.NewExporter(afero.NewOsFs()).ExportTo(ctx, analysis, opts.Export.JsonPath); err != nil {
			return fmt.Errorf("cannot export analysis: %w", err)
		}
		return nil
	}

	if opts.CI.Enabled {
		eval := adapter.NewEvaluator(opts.CI.Rules.List).Evaluate(ctx, analysis)

		if !eval.Pass {
			return errors.New("evaluation failed")
		}
		return nil
	}

	return explore(ctx, opts, ref)
}

// explore loads the image a second time through stereoscope for the TUI.
// todo: removed when commands are rewired
func explore(ctx context.Context, opts options.Application, ref string) (err error) {
	img, err := source.Load(ctx, ref, source.Config{Source: opts.Analysis.ContainerEngine, Platform: opts.Analysis.Platform})
	if err != nil {
		return fmt.Errorf("cannot load image: %w", err)
	}
	defer func() {
		err = errors.Join(err, img.Cleanup())
	}()

	analysis, err := dive.Analyze(ctx, ref, img)
	if err != nil {
		return fmt.Errorf("cannot analyze image: %w", err)
	}

	// blocks until the TUI returns, the image has to outlive it
	bus.ExploreAnalysis(ctx, analysis)
	return nil
}
