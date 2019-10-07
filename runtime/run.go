package runtime

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/runtime/ci"
	"github.com/wagoodman/dive/runtime/export"
	"os"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui"
	"github.com/wagoodman/dive/utils"
)

func runCi(analysis *image.AnalysisResult, options Options) {
	fmt.Printf("  efficiency: %2.4f %%\n", analysis.Efficiency*100)
	fmt.Printf("  wastedBytes: %d bytes (%s)\n", analysis.WastedBytes, humanize.Bytes(analysis.WastedBytes))
	fmt.Printf("  userWastedPercent: %2.4f %%\n", analysis.WastedUserPercent*100)

	evaluator := ci.NewCiEvaluator(options.CiConfig)

	pass := evaluator.Evaluate(analysis)
	evaluator.Report()

	if pass {
		os.Exit(0)
	}
	os.Exit(1)
}

func Run(options Options) {
	var err error
	doExport := options.ExportFile != ""
	doBuild := len(options.BuildArgs) > 0

	// if an image option was provided, parse it and determine the container image...
	// otherwise, use the configs default value.

	// if build is given, get the handler based off of either the explicit runtime

	imageResolver, err := dive.GetImageHandler(options.Engine)
	if err != nil {
		fmt.Printf("cannot determine image provider: %v\n", err)
		os.Exit(1)
	}

	var img *image.Image

	if doBuild {
		fmt.Println(utils.TitleFormat("Building image..."))
		img, err = imageResolver.Build(options.BuildArgs)
		if err != nil {
			fmt.Printf("cannot build image: %v\n", err)
			os.Exit(1)
		}
	} else {
		img, err = imageResolver.Fetch(options.ImageId)
		if err != nil {
			fmt.Printf("cannot fetch image: %v\n", err)
			os.Exit(1)
		}
	}

	// todo, cleanup on error
	// todo: image get should return error for cleanup?

	if doExport {
		fmt.Println(utils.TitleFormat(fmt.Sprintf("Analyzing image... (export to '%s')", options.ExportFile)))
	} else {
		fmt.Println(utils.TitleFormat("Analyzing image..."))
	}

	result, err := img.Analyze()
	if err != nil {
		fmt.Printf("cannot analyze image: %v\n", err)
		os.Exit(1)
	}

	if doExport {
		err = export.NewExport(result).ToFile(options.ExportFile)
		if err != nil {
			fmt.Printf("cannot write export file: %v\n", err)
			os.Exit(1)
		}
	}

	if options.Ci {
		runCi(result, options)
	} else {
		if doExport {
			os.Exit(0)
		}

		fmt.Println(utils.TitleFormat("Building cache..."))
		cache := filetree.NewFileTreeCache(result.RefTrees)
		err := cache.Build()
		if err != nil {
			logrus.Error(err)
			os.Exit(1)
		}

		// it appears there is a race condition where termbox.Init() will
		// block nearly indefinitely when running as the first process in
		// a Docker container when started within ~25ms of container startup.
		// I can't seem to determine the exact root cause, however, a large
		// enough sleep will prevent this behavior (todo: remove this hack)
		time.Sleep(100 * time.Millisecond)

		err = ui.Run(result, cache)
		if err != nil {
			logrus.Error(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}
