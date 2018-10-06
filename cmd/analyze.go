package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	"os"
	"github.com/wagoodman/dive/image"
	"github.com/wagoodman/dive/ui"
)

func analyze(cmd *cobra.Command, args []string) {
	userImage := args[0]
	if userImage == "" {
		fmt.Println("No image argument given")
		os.Exit(1)
	}
	manifest, refTrees := image.InitializeData(userImage)
	ui.Run(manifest, refTrees)
}