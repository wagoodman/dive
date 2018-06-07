package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
	"github.com/wagoodman/docker-image-explorer/image"
	"github.com/wagoodman/docker-image-explorer/ui"
)

const name = "die"
const version = "v0.0.0"
const author = "wagoodman"

func main() {
	app := cli.NewApp()
	app.Name = "die"
	app.Usage = "Explore your docker images"
	app.Action = func(c *cli.Context) error {
		userImage := c.Args().Get(0)
		if userImage == "" {
			fmt.Println("No image argument given")
			os.Exit(1)
		}
		manifest, refTrees := image.InitializeData(userImage)
		ui.Run(manifest, refTrees)
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
