package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/image"
	"github.com/wagoodman/dive/ui"
)

// analyze takes a docker image tag, digest, or id and displayes the
// image analysis to the screen
func analyze(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		printVersionFlag, err := cmd.PersistentFlags().GetBool("version")
		if err == nil && printVersionFlag {
			printVersion(cmd, args)
			return
		}

		fmt.Println("No image argument given")
		cmd.Help()
		os.Exit(1)
	}

	userImage := args[0]
	if userImage == "" {
		fmt.Println("No image argument given")
		cmd.Help()
		os.Exit(1)
	}
	color.New(color.Bold).Println("Analyzing Image")
	manifest, refTrees, efficiency, inefficiencies := image.InitializeData(userImage)
	ui.Run(manifest, refTrees, efficiency, inefficiencies)
}
