package runtime

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ci"
	"github.com/wagoodman/dive/runtime/export"
	"github.com/wagoodman/dive/runtime/ui"
	"github.com/wagoodman/dive/utils"
	"os"
	"time"
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
		events.message(utils.TitleFormat("Fetching image...") + " (this can take a while for large images)")
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

		evaluator := ci.NewCiEvaluator(options.CiConfig)
		pass := evaluator.Evaluate(analysis)
		events.message(evaluator.Report())

		if !pass {
			events.exitWithError(nil)
		}

		return

	} else {
		events.message(utils.TitleFormat("Building cache..."))
		treeStack := filetree.NewComparer(analysis.RefTrees)
		errors := treeStack.BuildCache()
		if errors != nil {
			for _, err := range errors {
				events.message("  " + err.Error())
			}
			if !options.IgnoreErrors {
				events.exitWithError(fmt.Errorf("file tree has path errors (use '--ignore-errors' to attempt to continue)"))
				return
			}
		}

		if enableUi {
			// it appears there is a race condition where termbox.Init() will
			// block nearly indefinitely when running as the first process in
			// a Docker container when started within ~25ms of container startup.
			// I can't seem to determine the exact root cause, however, a large
			// enough sleep will prevent this behavior (todo: remove this hack)
			time.Sleep(100 * time.Millisecond)

			err = ui.Run(analysis, treeStack)
			if err != nil {
				events.exitWithError(err)
				return
			}
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

	for event := range events {
		if event.stdout != "" {
			fmt.Println(event.stdout)
		}

		if event.stderr != "" {
			_, err := fmt.Fprintln(os.Stderr, event.stderr)
			if err != nil {
				fmt.Println("error: could not write to buffer:", err)
			}
		}

		if event.err != nil {
			logrus.Error(event.err)
			_, err := fmt.Fprintln(os.Stderr, event.err.Error())
			if err != nil {
				fmt.Println("error: could not write to buffer:", err)
			}
		}

		if event.errorOnExit {
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}
