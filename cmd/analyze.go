package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/image"
	"github.com/wagoodman/dive/ui"
	"github.com/wagoodman/dive/utils"
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
	ui.Run(manifest, refTrees, efficiency, inefficiencies)
}
