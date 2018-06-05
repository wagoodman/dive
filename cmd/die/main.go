package main

import (
	"os"
	"github.com/wagoodman/docker-image-explorer/image"
	"github.com/wagoodman/docker-image-explorer/ui"
)

const name = "die"
const version = "v0.0.0"
const author = "wagoodman"

func main() {
	os.Exit(run(os.Args))
}


func run(args []string) int {
	image.WriteImage()
	manifest, refTrees := image.InitializeData()

	ui.Run(manifest, refTrees)
	return 0
}

