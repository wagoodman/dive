package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/runtime"
	"github.com/wagoodman/dive/utils"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:                "build [any valid `docker build` arguments]",
	Short:              "Builds and analyzes a docker image from a Dockerfile (this is a thin wrapper for the `docker build` command).",
	DisableFlagParsing: true,
	Run:                doBuildCmd,
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

// doBuildCmd implements the steps taken for the build command
func doBuildCmd(cmd *cobra.Command, args []string) {
	defer utils.Cleanup()

	initLogging()

	engine, err := cmd.PersistentFlags().GetString("engine")
	if err != nil {
		fmt.Printf("unable to determine eingine: %v\n", err)
		utils.Exit(1)
	}

	runtime.Run(runtime.Options{
		Ci:         isCi,
		Engine:     dive.GetEngine(engine),
		BuildArgs:  args,
		ExportFile: exportFile,
		CiConfig:   ciConfig,
	})
}
