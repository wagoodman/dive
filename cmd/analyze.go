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

// analyze takes a docker image tag, digest, or id and displays the
// image analysis to the screen
func analyze(cmd *cobra.Command, args []string) {
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
	color.New(color.Bold).Println("Analyzing Image")
	manifest, refTrees, efficiency, inefficiencies := image.InitializeData(userImage)

	if exportFile != "" {
		// todo: support statistics
		exportStatistics(manifest, refTrees, efficiency, inefficiencies)
		color.New(color.Bold).Println(fmt.Sprintf("Exported to %s", exportFile))
	} else {
		ui.Run(manifest, refTrees, efficiency, inefficiencies)
	}
}

type export struct {
	Layer []exportLayer `json:"layer"`
	Image exportImage   `json:"image"`
}

type exportLayer struct {
	DigestID  string `json:"digestId"`
	TarID     string `json:"tarId"`
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

func newExport(layers []*image.Layer, refTrees []*filetree.FileTree, efficiency float64, inefficiencies filetree.EfficiencySlice) *export {
	data := export{}
	data.Layer = make([]exportLayer, len(layers))
	data.Image.InefficientFiles = make([]inefficientFiles, len(inefficiencies))

	// export layers in order
	for revIdx := len(layers) - 1; revIdx >= 0; revIdx-- {
		layer := layers[revIdx]
		idx := (len(layers) - 1) - revIdx

		data.Layer[idx] = exportLayer{
			DigestID:  layer.History.ID,
			TarID:     layer.TarId(),
			SizeBytes: layer.History.Size,
			Command:   layer.History.CreatedBy,
		}
	}

	// export image info
	data.Image.SizeBytes = 0
	for idx := 0; idx < len(layers); idx++ {
		data.Image.SizeBytes += layers[idx].History.Size
	}

	data.Image.EfficiencyScore = efficiency

	for idx := 0; idx < len(inefficiencies); idx++ {
		fileData := inefficiencies[len(inefficiencies)-1-idx]
		data.Image.InefficientBytes += uint64(fileData.CumulativeSize)

		data.Image.InefficientFiles[idx] = inefficientFiles{
			Count:     len(fileData.Nodes),
			SizeBytes: uint64(fileData.CumulativeSize),
			File:      fileData.Path,
		}
	}

	return &data
}

func exportStatistics(layers []*image.Layer, refTrees []*filetree.FileTree, efficiency float64, inefficiencies filetree.EfficiencySlice) {
	data := newExport(layers, refTrees, efficiency, inefficiencies)
	payload, err := json.MarshalIndent(&data, "", "  ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(exportFile, payload, 0644)
	if err != nil {
		panic(err)
	}
}
