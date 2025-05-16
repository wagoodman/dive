package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/anchore/clio"
	"github.com/spf13/afero"
	adapterV1 "github.com/wagoodman/dive/cmd/dive/cli/internal/command/adapter/v1"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui"
	diveV1 "github.com/wagoodman/dive/dive/v1"
	imageV1 "github.com/wagoodman/dive/dive/v1/image"
	"github.com/wagoodman/dive/internal/bus"
	"os"
)

func v1BuildImage(ctx context.Context, opts options.Application, app clio.Application, args []string) error {
	if err := setV1UI(app, opts); err != nil {
		return fmt.Errorf("failed to set UI: %w", err)
	}

	resolver, err := diveV1.GetImageResolver(opts.Analysis.Source)
	if err != nil {
		return fmt.Errorf("cannot determine image provider for build: %w", err)
	}

	img, err := adapterV1.ImageResolver(resolver).Build(ctx, args)
	if err != nil {
		return fmt.Errorf("cannot build image: %w", err)
	}

	return runV1(ctx, opts, img, resolver)
}

func v1FetchImage(ctx context.Context, opts options.Application, app clio.Application) error {
	if err := setV1UI(app, opts); err != nil {
		return fmt.Errorf("failed to set UI: %w", err)
	}

	resolver, err := diveV1.GetImageResolver(opts.Analysis.Source)
	if err != nil {
		return fmt.Errorf("cannot determine image provider to fetch from: %w", err)
	}

	img, err := adapterV1.ImageResolver(resolver).Fetch(ctx, opts.Analysis.Image)
	if err != nil {
		return fmt.Errorf("cannot load image: %w", err)
	}

	return runV1(ctx, opts, img, resolver)
}

func setV1UI(app clio.Application, opts options.Application) error {
	type Stater interface {
		State() *clio.State
	}

	state := app.(Stater).State()

	ux := ui.NewV1UI(opts.V1Preferences(), os.Stdout, state.Config.Log.Quiet, state.Config.Log.Verbosity)
	return state.UI.Replace(ux)
}

func runV1(ctx context.Context, opts options.Application, img *imageV1.Image, content imageV1.ContentReader) error {
	analysis, err := adapterV1.NewAnalyzer().Analyze(ctx, img)
	if err != nil {
		return fmt.Errorf("cannot analyze image: %w", err)
	}

	if opts.Export.JsonPath != "" {
		if err := adapterV1.NewExporter(afero.NewOsFs()).ExportTo(ctx, analysis, opts.Export.JsonPath); err != nil {
			return fmt.Errorf("cannot export analysis: %w", err)
		}
		return nil
	}

	if opts.CI.Enabled {
		eval := adapterV1.NewEvaluator(opts.CI.Rules.List).Evaluate(ctx, analysis)

		if !eval.Pass {
			return errors.New("evaluation failed")
		}
		return nil
	}

	bus.ExploreAnalysis(*analysis, content)

	return nil
}
