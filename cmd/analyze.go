package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/runtime"
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

	initLogging()

	runtime.Run(runtime.Options{
		ImageId:      userImage,
		ExportFile:   exportFile,
		CiConfigFile: ciConfigFile,
	})
}
