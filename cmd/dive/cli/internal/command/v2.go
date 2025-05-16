package command

import (
	"context"
	"fmt"
	"github.com/anchore/clio"
	"github.com/anchore/stereoscope"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui"
	diveV1 "github.com/wagoodman/dive/dive/v1"
	"os"
)

//func v2BuildImage(ctx context.Context, opts options.Application, app clio.Application, args []string) error {
//	if err := setV1UI(app, opts); err != nil {
//		return fmt.Errorf("failed to set UI: %w", err)
//	}
//
//	resolver, err := diveV1.GetImageResolver(opts.Analysis.Source)
//	if err != nil {
//		return fmt.Errorf("cannot determine image provider for build: %w", err)
//	}
//
//	img, err := adapterV1.ImageResolver(resolver).Build(ctx, args)
//	if err != nil {
//		return fmt.Errorf("cannot build image: %w", err)
//	}
//
//	return runV1(ctx, opts, img, resolver)
//}

func v2FetchImage(ctx context.Context, opts options.Application, app clio.Application) error {
	if err := setV2UI(app, opts); err != nil {
		return fmt.Errorf("failed to set UI: %w", err)
	}
	imageStr := opts.Analysis.Image
	var src string
	switch opts.Analysis.Source {
	case diveV1.SourceDockerEngine:
		src = "docker"
	case diveV1.SourcePodmanEngine:
		src = "podman"
	case diveV1.SourceDockerArchive:
		src = "docker-archive"
	}

	if src != "" {
		imageStr = fmt.Sprintf("%s:%s", src, imageStr)
	}

	image, err := stereoscope.GetImage(ctx, imageStr)
	if err != nil {
		return fmt.Errorf("cannot load %q: %w", imageStr, err)
	}

	// note: we are writing out temp files which should be cleaned up after you're done with the image object
	defer image.Cleanup()

	// TODO!
	//return runV2(ctx, opts, img, resolver)
	return nil
}

func setV2UI(app clio.Application, opts options.Application) error {
	type Stater interface {
		State() *clio.State
	}

	state := app.(Stater).State()

	// TODO: opts.V1Preferences(),
	ux := ui.NewV2UI(app, os.Stdout, state.Config.Log.Quiet, state.Config.Log.Verbosity)
	return state.UI.Replace(ux)
}

//func runV2(ctx context.Context, opts options.Application, img *imageV1.Image, content imageV1.ContentReader) error {
//	analysis, err := adapterV1.NewAnalyzer().Analyze(ctx, img)
//	if err != nil {
//		return fmt.Errorf("cannot analyze image: %w", err)
//	}
//
//	if opts.Export.JsonPath != "" {
//		if err := adapterV1.NewExporter(afero.NewOsFs()).ExportTo(ctx, analysis, opts.Export.JsonPath); err != nil {
//			return fmt.Errorf("cannot export analysis: %w", err)
//		}
//		return nil
//	}
//
//	if opts.CI.Enabled {
//		eval := adapterV1.NewEvaluator(opts.CI.Rules.List).Evaluate(ctx, analysis)
//
//		if !eval.Pass {
//			return errors.New("evaluation failed")
//		}
//		return nil
//	}
//
//	bus.ExploreAnalysis(*analysis, content)
//
//	return nil
//}
