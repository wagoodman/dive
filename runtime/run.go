package runtime

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/logrusorgru/aurora"
	"github.com/wagoodman/dive/filetree"
	"github.com/wagoodman/dive/image"
	"github.com/wagoodman/dive/runtime/ci"
	"github.com/wagoodman/dive/ui"
	"github.com/wagoodman/dive/utils"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func title(s string) string {
	return aurora.Bold(s).String()
}

func runCi(analysis *image.AnalysisResult, options Options) {
	fmt.Printf("  efficiency: %2.4f %%\n", analysis.Efficiency*100)
	fmt.Printf("  wastedBytes: %d bytes (%s)\n", analysis.WastedBytes, humanize.Bytes(analysis.WastedBytes))
	fmt.Printf("  userWastedPercent: %2.4f %%\n", analysis.WastedUserPercent*100)

	fmt.Println(title("Run CI Validations..."))
	evaluator := ci.NewEvaluator()

	err := evaluator.LoadConfig(options.CiConfigFile)
	if err != nil {
		fmt.Println("  Using default CI config")
	} else {
		fmt.Printf("  Using CI config: %s\n", options.CiConfigFile)
	}

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
	isCi, _ := strconv.ParseBool(os.Getenv("CI"))

	if doBuild {
		fmt.Println(title("Building image..."))
		options.ImageId = runBuild(options.BuildArgs)
	}

	analyzer := image.GetAnalyzer(options.ImageId)

	fmt.Println(title("Fetching image...") + " (this can take a while with large images)")
	reader, err := analyzer.Fetch()
	if err != nil {
		fmt.Printf("cannot fetch image: %v\n", err)
		utils.Exit(1)
	}
	defer reader.Close()

	fmt.Println(title("Parsing image..."))
	err = analyzer.Parse(reader)
	if err != nil {
		fmt.Printf("cannot parse image: %v\n", err)
		utils.Exit(1)
	}

	if doExport {
		fmt.Println(title(fmt.Sprintf("Analyzing image... (export to '%s')", options.ExportFile)))
	} else {
		fmt.Println(title("Analyzing image..."))
	}

	result, err := analyzer.Analyze()
	if err != nil {
		fmt.Printf("cannot analyze image: %v\n", err)
		utils.Exit(1)
	}

	if doExport {
		err = newExport(result).toFile(options.ExportFile)
		if err != nil {
			fmt.Printf("cannot write export file: %v\n", err)
			utils.Exit(1)
		}
	}

	if isCi {
		runCi(result, options)
	} else {
		if doExport {
			utils.Exit(0)
		}

		fmt.Println(title("Building cache..."))
		cache := filetree.NewFileTreeCache(result.RefTrees)
		cache.Build()

		ui.Run(result, cache)
	}
}
