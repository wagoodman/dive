package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/utils"
	"io/ioutil"
	"os"
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
	iidfile, err := ioutil.TempFile("/tmp", "dive.*.iid")
	if err != nil {
		utils.Cleanup()
		log.Fatal(err)
	}
	defer os.Remove(iidfile.Name())

	allArgs := append([]string{"--iidfile", iidfile.Name()}, args...)
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

	run(string(imageId))
}
