package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/image"
	"github.com/wagoodman/dive/ui"
	"github.com/wagoodman/dive/utils"
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
	color.New(color.Bold).Println("Analyzing Image")

	ui.Run(fetchAndAnalyze(userImage))
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
