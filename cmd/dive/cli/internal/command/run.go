package command

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/anchore/clio"
	"github.com/anchore/stereoscope/pkg/filetree/filenode"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/dustin/go-humanize"

	"github.com/wagoodman/dive/cmd/dive/cli/internal/command/ci"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/command/export"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/dive/source"
	"github.com/wagoodman/dive/internal/bus"
	"github.com/wagoodman/dive/internal/bus/event/payload"
	"github.com/wagoodman/dive/internal/log"
)

func setUI(app clio.Application, opts options.Application, ux func(options.Application, *clio.State) clio.UI) error {
	type Stater interface {
		State() *clio.State
	}

	state := app.(Stater).State()
	return state.UI.Replace(ux(opts, state))
}

// progressUI is the bubbletea progress UI where the terminal allows it, the line UI otherwise
func progressUI(opts options.Application, state *clio.State) clio.UI {
	return ui.New(opts.TUIPreferences(), os.Stdout, state.Config.Log.Quiet, state.Config.Log.Verbosity)
}

// lineUI never takes over the terminal, for commands that share it with a subprocess (dive build)
func lineUI(opts options.Application, state *clio.State) clio.UI {
	return ui.NewLineUI(opts.TUIPreferences(), os.Stdout, state.Config.Log.Quiet, state.Config.Log.Verbosity)
}

// run loads the image behind input, analyzes it, then exports the analysis (--json), evaluates it (--ci) or hands
// it to the TUI.
func run(ctx context.Context, opts options.Application, input string) (err error) {
	_, ref := source.Parse(input, opts.Analysis.ContainerEngine)

	img, err := load(ctx, input, ref, opts.Analysis)
	if err != nil {
		return fmt.Errorf("cannot load image: %w", err)
	}
	defer func() {
		// interrupts skip this, source.Cleanup (called from main) covers that path
		err = errors.Join(err, img.Cleanup())
	}()

	analysis, err := analyze(ctx, ref, img)
	if err != nil {
		return fmt.Errorf("cannot analyze image: %w", err)
	}

	switch {
	case opts.Export.JsonPath != "":
		if err := exportTo(analysis, opts.Export.JsonPath); err != nil {
			return fmt.Errorf("cannot export analysis: %w", err)
		}
		return nil
	case opts.CI.Enabled:
		if !evaluate(ctx, analysis, opts.CI.Rules.List) {
			return errors.New("evaluation failed")
		}
		return nil
	}

	// blocks until the TUI returns, the image has to outlive it
	bus.ExploreAnalysis(ctx, analysis)
	return nil
}

func load(ctx context.Context, input, ref string, opts options.Analysis) (*image.Image, error) {
	log.WithFields("image", ref).Info("fetching")
	mon := bus.StartTask(payload.GenericTask{
		Title: payload.Title{
			Default:      "Loading image",
			WhileRunning: "Loading image",
			OnSuccess:    "Fetched image",
		},
		ID:      ref,
		Context: ref,
	})

	img, err := source.Load(ctx, input, source.Config{Source: opts.ContainerEngine, Platform: opts.Platform})
	finish(mon, err)
	return img, err
}

func analyze(ctx context.Context, ref string, img *image.Image) (*dive.Analysis, error) {
	log.WithFields("image", ref).Infof("analyzing")

	var files int
	var fileSize uint64
	for _, l := range img.Layers {
		fileSize += uint64(l.Metadata.Size)
		_ = dive.WalkTree(l.Tree, func(*filenode.FileNode, bool) error {
			files++
			return nil
		})
	}
	fileSizeStr := humanize.Bytes(fileSize)
	filesStr := humanize.Comma(int64(files))

	log.Debugf("├── layers: %d", len(img.Layers))
	log.Debugf("├── files: %s", filesStr)
	log.Debugf("└── file size: %s", fileSizeStr)

	mon := bus.StartTask(payload.GenericTask{
		Title: payload.Title{
			Default:      "Analyzing image",
			WhileRunning: "Analyzing image",
			OnSuccess:    "Analyzed image",
		},
		ID:      ref,
		Context: fmt.Sprintf("[layers:%d files:%s size:%s]", len(img.Layers), filesStr, fileSizeStr),
	})

	analysis, err := dive.Analyze(ctx, ref, img)
	finish(mon, err)
	return analysis, err
}

func exportTo(analysis *dive.Analysis, path string) error {
	log.WithFields("path", path).Infof("exporting analysis")
	mon := bus.StartTask(payload.GenericTask{
		Title: payload.Title{
			Default:      "Exporting details",
			WhileRunning: "Exporting details",
			OnSuccess:    "Exported details",
		},
		ID:      analysis.Ref,
		Context: fmt.Sprintf("[file: %s]", path),
	})

	data, err := export.NewExport(analysis).Marshal()
	if err == nil {
		err = os.WriteFile(path, data, 0o644)
	}
	finish(mon, err)
	return err
}

// evaluate runs the CI rules and publishes the report, returning whether the image passed
func evaluate(ctx context.Context, analysis *dive.Analysis, rules []ci.Rule) bool {
	log.WithFields("image", analysis.Ref).Infof("evaluating image")
	mon := bus.StartTask(payload.GenericTask{
		Title: payload.Title{
			Default:      "Evaluating image",
			WhileRunning: "Evaluating image",
			OnSuccess:    "Evaluated image",
		},
		ID:      analysis.Ref,
		Context: fmt.Sprintf("[rules: %d]", len(rules)),
	})

	eval := ci.NewEvaluator(rules).Evaluate(ctx, ci.NewStats(analysis))
	if eval.Pass {
		finish(mon, nil)
	} else {
		finish(mon, fmt.Errorf("failed evaluation"))
	}
	bus.Report(eval.Report)
	return eval.Pass
}

// finish marks a task row done, failed when err is set. SetCompleted would replace the error, so a failed task is
// completed through its size instead (otherwise its spinner never stops).
func finish(mon *payload.GenericProgress, err error) {
	if err == nil {
		mon.SetCompleted()
		return
	}
	mon.SetError(err)
	mon.Manual.SetTotal(1)
	mon.Manual.Set(1)
}
