package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/filetree"
	"github.com/wagoodman/dive/image"
	"github.com/wagoodman/dive/ui"
	"github.com/wagoodman/dive/utils"
	"io/ioutil"
)

// doAnalyzeCmd takes a docker image tag, digest, or id and displays the
// image analysis to the screen
func doAnalyzeCmd(cmd *cobra.Command, args []string) {
	defer utils.Cleanup()
	if len(args) == 0 {
		printVersionFlag, err := cmd.PersistentFlags().GetBool("version")
		if err == nil && printVersionFlag {
			printVersion(cmd, args)
			return
		}

		fmt.Println("No image argument given")
		cmd.Help()
		utils.Exit(1)
	}

	userImage := args[0]
	if userImage == "" {
		fmt.Println("No image argument given")
		cmd.Help()
		utils.Exit(1)
	}

	run(userImage)
}

type export struct {
	Layer []exportLayer `json:"layer"`
	Image exportImage   `json:"image"`
}

type exportLayer struct {
	Index     int    `json:"index"`
	DigestID  string `json:"digestId"`
	SizeBytes uint64 `json:"sizeBytes"`
	Command   string `json:"command"`
}
type exportImage struct {
	SizeBytes        uint64             `json:"sizeBytes"`
	InefficientBytes uint64             `json:"inefficientBytes"`
	EfficiencyScore  float64            `json:"efficiencyScore"`
	InefficientFiles []inefficientFiles `json:"inefficientFiles"`
}

type inefficientFiles struct {
	Count     int    `json:"count"`
	SizeBytes uint64 `json:"sizeBytes"`
	File      string `json:"file"`
}

func newExport(analysis *image.AnalysisResult) *export {
	data := export{}
	data.Layer = make([]exportLayer, len(analysis.Layers))
	data.Image.InefficientFiles = make([]inefficientFiles, len(analysis.Inefficiencies))

	// export layers in order
	for revIdx := len(analysis.Layers) - 1; revIdx >= 0; revIdx-- {
		layer := analysis.Layers[revIdx]
		idx := (len(analysis.Layers) - 1) - revIdx

		data.Layer[idx] = exportLayer{
			Index:     idx,
			DigestID:  layer.Id(),
			SizeBytes: layer.Size(),
			Command:   layer.Command(),
		}
	}

	// export image info
	data.Image.SizeBytes = 0
	for idx := 0; idx < len(analysis.Layers); idx++ {
		data.Image.SizeBytes += analysis.Layers[idx].Size()
	}

	data.Image.EfficiencyScore = analysis.Efficiency

	for idx := 0; idx < len(analysis.Inefficiencies); idx++ {
		fileData := analysis.Inefficiencies[len(analysis.Inefficiencies)-1-idx]
		data.Image.InefficientBytes += uint64(fileData.CumulativeSize)

		data.Image.InefficientFiles[idx] = inefficientFiles{
			Count:     len(fileData.Nodes),
			SizeBytes: uint64(fileData.CumulativeSize),
			File:      fileData.Path,
		}
	}

	return &data
}

func exportStatistics(analysis *image.AnalysisResult) {
	data := newExport(analysis)
	payload, err := json.MarshalIndent(&data, "", "  ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(exportFile, payload, 0644)
	if err != nil {
		panic(err)
	}
}

func fetchAndAnalyze(imageID string) *image.AnalysisResult {
	analyzer := image.GetAnalyzer(imageID)

	fmt.Println("  Fetching image...")
	err := analyzer.Parse(imageID)
	if err != nil {
		fmt.Printf("cannot fetch image: %v\n", err)
		utils.Exit(1)
	}

	fmt.Println("  Analyzing image...")
	result, err := analyzer.Analyze()
	if err != nil {
		fmt.Printf("cannot doAnalyzeCmd image: %v\n", err)
		utils.Exit(1)
	}
	return result
}

func run(imageID string) {
	color.New(color.Bold).Println("Analyzing Image")
	result := fetchAndAnalyze(imageID)

	if exportFile != "" {
		exportStatistics(result)
		color.New(color.Bold).Println(fmt.Sprintf("Exported to %s", exportFile))
		utils.Exit(0)
	}

	fmt.Println("  Building cache...")
	cache := filetree.NewFileTreeCache(result.RefTrees)
	cache.Build()

	ui.Run(result, cache)
}
