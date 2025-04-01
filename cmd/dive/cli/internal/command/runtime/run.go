package runtime

import (
	"errors"
	"fmt"
	"github.com/spf13/afero"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/command/runtime/ci"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/internal/bus"
	"golang.org/x/net/context"
)

type Config struct {
	// request
	Image     string
	Source    dive.ImageSource
	BuildArgs []string

	// gating
	Ci         bool
	CiRules    []ci.Rule
	ExportFile string

	// ui
	UI v1.Preferences
}

func Run(ctx context.Context, cfg Config) error {

	imageResolver, err := dive.GetImageResolver(cfg.Source)
	if err != nil {
		return errors.New("cannot determine image provider")
	}

	ir := defaultImageResolver(cfg, imageResolver)
	analyzer := defaultAnalyzer()
	exporter := defaultExporter(afero.NewOsFs())
	evaluator := defaultEvaluator(cfg.CiRules)

	return run(ctx, true, cfg, ir, analyzer, exporter, evaluator)
}

func run(ctx context.Context, enableUI bool, cfg Config, imageResolver ImageResolver, analyzer Analyzer, exporter Exporter, evaluator Evaluator) error {
	doExport := cfg.ExportFile != ""

	img, err := imageResolver.Get(ctx)
	if err != nil {
		return err
	}

	analysis, err := analyzer.Analyze(ctx, img)
	if err != nil {
		return fmt.Errorf("cannot analyze image: %w", err)
	}

	if doExport {
		if err := exporter.ExportTo(ctx, analysis, cfg.ExportFile); err != nil {
			return fmt.Errorf("cannot export analysis: %w", err)
		}
		return nil
	}

	if cfg.Ci {
		eval := evaluator.Evaluate(ctx, analysis)

		if !eval.Pass {
			return errors.New("evaluation failed")
		}
		return nil
	}

	if enableUI {
		bus.ExploreAnalysis(*analysis, imageResolver)
	}
	return nil
}
