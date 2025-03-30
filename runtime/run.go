package runtime

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	v1 "github.com/wagoodman/dive/runtime/ui/v1"
	"github.com/wagoodman/dive/runtime/ui/v1/app"
	"os"

	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ci"
	"github.com/wagoodman/dive/runtime/export"
	"github.com/wagoodman/dive/utils"
)

func run(enableUi bool, options Options, imageResolver image.Resolver, events eventChannel, filesystem afero.Fs) {
	var img *image.Image
	var err error
	defer close(events)

	doExport := options.ExportFile != ""
	doBuild := len(options.BuildArgs) > 0

	if doBuild {
		events.message(utils.TitleFormat("Building image..."))
		img, err = imageResolver.Build(options.BuildArgs)
		if err != nil {
			events.exitWithErrorMessage("cannot build image", err)
			return
		}
	} else {
		events.message(utils.TitleFormat("Image Source: ") + options.Source.String() + "://" + options.Image)
		events.message(utils.TitleFormat("Extracting image from "+imageResolver.Name()+"...") + " (this can take a while for large images)")
		img, err = imageResolver.Fetch(options.Image)
		if err != nil {
			events.exitWithErrorMessage("cannot fetch image", err)
			return
		}
	}

	events.message(utils.TitleFormat("Analyzing image..."))
	analysis, err := img.Analyze()
	if err != nil {
		events.exitWithErrorMessage("cannot analyze image", err)
		return
	}

	if doExport {
		events.message(utils.TitleFormat(fmt.Sprintf("Exporting image to '%s'...", options.ExportFile)))
		bytes, err := export.NewExport(analysis).Marshal()
		if err != nil {
			events.exitWithErrorMessage("cannot marshal export payload", err)
			return
		}

		file, err := filesystem.OpenFile(options.ExportFile, os.O_RDWR|os.O_CREATE, 0644)
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

	if options.Ci {
		events.message(fmt.Sprintf("  efficiency: %2.4f %%", analysis.Efficiency*100))
		events.message(fmt.Sprintf("  wastedBytes: %d bytes (%s)", analysis.WastedBytes, humanize.Bytes(analysis.WastedBytes)))
		events.message(fmt.Sprintf("  userWastedPercent: %2.4f %%", analysis.WastedUserPercent*100))

		evaluator := ci.NewEvaluator(options.CiRules)
		pass := evaluator.Evaluate(analysis)
		events.message(evaluator.Report())

		if !pass {
			events.exitWithError(nil)
		}

		return
	} else if enableUi {
		err = app.Run(v1.Config{
			Image:        options.Image,
			Content:      imageResolver,
			Analysis:     analysis,
			KeyBindings:  options.KeyBindings,
			IgnoreErrors: options.IgnoreErrors,
		})
		if err != nil {
			events.exitWithError(err)
			return
		}
	}
}

func Run(options Options) {
	var exitCode int
	var events = make(eventChannel)

	imageResolver, err := dive.GetImageResolver(options.Source)
	if err != nil {
		message := "cannot determine image provider"
		logrus.Error(message)
		logrus.Error(err)
		fmt.Fprintf(os.Stderr, "%s: %+v\n", message, err)
		os.Exit(1)
	}

	go run(true, options, imageResolver, events, afero.NewOsFs())

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
			logrus.Error(e.err)
			_, err := fmt.Fprintln(os.Stderr, e.err.Error())
			if err != nil {
				fmt.Println("error: could not write to buffer:", err)
			}
		}

		if e.errorOnExit {
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}
