package runtime

import (
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/spf13/afero"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/command/runtime/ci"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/command/runtime/export"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/app"
	"github.com/wagoodman/dive/internal/utils"
	"golang.org/x/net/context"
	"os"

	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/dive/image"
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
	var events = make(eventChannel)

	imageResolver, err := dive.GetImageResolver(cfg.Source)
	if err != nil {
		return errors.New("cannot determine image provider")
	}

	go run(ctx, true, cfg, imageResolver, events, afero.NewOsFs())

	var retErr error
	for e := range events {
		if e.stdout != "" {
			fmt.Println(e.stdout)
		}

		if e.stderr != "" {
			_, err := fmt.Fprintln(os.Stderr, e.stderr)
			if err != nil {
				fmt.Println("error: could not write to buffer:", err)
			}
		}

		if e.err != nil {
			retErr = errors.Join(retErr, e.err)
		}
	}

	return retErr
}

func run(ctx context.Context, enableUI bool, cfg Config, imageResolver image.Resolver, events eventChannel, filesystem afero.Fs) {
	var img *image.Image
	var err error
	defer close(events)

	doExport := cfg.ExportFile != ""
	doBuild := len(cfg.BuildArgs) > 0

	if doBuild {
		events.message(utils.TitleFormat("Building image..."))
		img, err = imageResolver.Build(ctx, cfg.BuildArgs)
		if err != nil {
			events.exitWithErrorMessage("cannot build image", err)
			return
		}
	} else {
		events.message(utils.TitleFormat("Image Source: ") + cfg.Source.String() + "://" + cfg.Image)
		events.message(utils.TitleFormat("Extracting image from "+imageResolver.Name()+"...") + " (this can take a while for large images)")
		img, err = imageResolver.Fetch(ctx, cfg.Image)
		if err != nil {
			events.exitWithErrorMessage("cannot fetch image", err)
			return
		}
	}

	events.message(utils.TitleFormat("Analyzing image..."))
	analysis, err := img.Analyze(ctx)
	if err != nil {
		events.exitWithErrorMessage("cannot analyze image", err)
		return
	}
	if analysis == nil {
		events.exitWithErrorMessage("cannot analyze image", fmt.Errorf("no results returned"))
		return
	}

	if doExport {
		events.message(utils.TitleFormat(fmt.Sprintf("Exporting image to '%s'...", cfg.ExportFile)))
		bytes, err := export.NewExport(analysis).Marshal()
		if err != nil {
			events.exitWithErrorMessage("cannot marshal export payload", err)
			return
		}

		file, err := filesystem.OpenFile(cfg.ExportFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			events.exitWithErrorMessage("cannot open export file", err)
			return
		}
		defer file.Close()

		_, err = file.Write(bytes)
		if err != nil {
			events.exitWithErrorMessage("cannot write to export file", err)
		}
		return
	}

	if cfg.Ci {
		events.message(fmt.Sprintf("  efficiency: %2.4f %%", analysis.Efficiency*100))
		events.message(fmt.Sprintf("  wastedBytes: %d bytes (%s)", analysis.WastedBytes, humanize.Bytes(analysis.WastedBytes)))
		events.message(fmt.Sprintf("  userWastedPercent: %2.4f %%", analysis.WastedUserPercent*100))

		evaluator := ci.NewEvaluator(cfg.CiRules)
		pass := evaluator.Evaluate(ctx, analysis)
		events.message(evaluator.Report())

		if !pass {
			events.exitWithError(nil)
		}

		return
	}

	if enableUI {
		err = app.Run(
			ctx,
			v1.Config{
				Content:     imageResolver,
				Analysis:    *analysis,
				Preferences: cfg.UI,
			},
		)
		if err != nil {
			events.exitWithError(err)
		}
	}
}

type eventChannel chan event

type event struct {
	stdout      string
	stderr      string
	err         error
	errorOnExit bool
}

func (ec eventChannel) message(msg string) {
	ec <- event{
		stdout: msg,
	}
}

func (ec eventChannel) exitWithError(err error) {
	ec <- event{
		err:         err,
		errorOnExit: true,
	}
}

func (ec eventChannel) exitWithErrorMessage(msg string, err error) {
	ec <- event{
		stderr:      msg,
		err:         err,
		errorOnExit: true,
	}
}
