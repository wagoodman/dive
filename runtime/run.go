package runtime

import (
	"fmt"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/runtime/ci"
	"github.com/wagoodman/dive/runtime/export"
	"io/ioutil"
	"log"
	"os"

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
		utils.Exit(0)
	}
	utils.Exit(1)
}

func runBuild(buildArgs []string) string {
	iidfile, err := ioutil.TempFile("/tmp", "dive.*.iid")
	if err != nil {
		utils.Cleanup()
		log.Fatal(err)
	}
	defer os.Remove(iidfile.Name())

	allArgs := append([]string{"--iidfile", iidfile.Name()}, buildArgs...)
	err = utils.RunDockerCmd("build", allArgs...)
	if err != nil {
		utils.Cleanup()
		log.Fatal(err)
	}

	imageId, err := ioutil.ReadFile(iidfile.Name())
	if err != nil {
		utils.Cleanup()
		log.Fatal(err)
	}

	return string(imageId)
}

func Run(options Options) {
	doExport := options.ExportFile != ""
	doBuild := len(options.BuildArgs) > 0

	if doBuild {
		fmt.Println(utils.TitleFormat("Building image..."))
		options.ImageId = runBuild(options.BuildArgs)
	}

	analyzer := dive.GetAnalyzer(options.ImageId)

	fmt.Println(utils.TitleFormat("Fetching image...") + " (this can take a while with large images)")
	reader, err := analyzer.Fetch()
	if err != nil {
		fmt.Printf("cannot fetch image: %v\n", err)
		utils.Exit(1)
	}
	defer reader.Close()

	fmt.Println(utils.TitleFormat("Parsing image..."))
	err = analyzer.Parse(reader)
	if err != nil {
		fmt.Printf("cannot parse image: %v\n", err)
		utils.Exit(1)
	}

	if doExport {
		fmt.Println(utils.TitleFormat(fmt.Sprintf("Analyzing image... (export to '%s')", options.ExportFile)))
	} else {
		fmt.Println(utils.TitleFormat("Analyzing image..."))
	}

	result, err := analyzer.Analyze()
	if err != nil {
		fmt.Printf("cannot analyze image: %v\n", err)
		utils.Exit(1)
	}

	if doExport {
		err = export.NewExport(result).ToFile(options.ExportFile)
		if err != nil {
			fmt.Printf("cannot write export file: %v\n", err)
			utils.Exit(1)
		}
	}

	if options.Ci {
		runCi(result, options)
	} else {
		if doExport {
			utils.Exit(0)
		}

		fmt.Println(utils.TitleFormat("Building cache..."))
		cache := filetree.NewFileTreeCache(result.RefTrees)
		cache.Build()

		ui.Run(result, cache)
	}
}
