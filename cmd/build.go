package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/runtime"
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
	initLogging()

	// there is no cli options allowed, only config can be supplied
	// todo: allow for an engine flag to be passed to dive but not the container engine
	engine := viper.GetString("container-engine")

	runtime.Run(runtime.Options{
		Ci:         isCi,
		Source:     dive.ParseImageSource(engine),
		BuildArgs:  args,
		ExportFile: exportFile,
		CiConfig:   ciConfig,
	})
}
