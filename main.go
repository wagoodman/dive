package main

import (
	"encoding/json"
	"fmt"

	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	// imageID := "golang:alpine"
	imageID := "die-test:latest"

	for {
		inspect, _, err := cli.ImageInspectWithRaw(ctx, imageID)
		if err != nil {
			panic(err)
		}

		history, err := cli.ImageHistory(ctx, imageID)
		if err != nil {
			panic(err)
		}

		historyStr, err := json.MarshalIndent(history, "", "  ")
		if err != nil {
			panic(err)
		}

		layerStr := ""
		for idx, layer := range inspect.RootFS.Layers {
			prefix := "├── "
			if idx == len(inspect.RootFS.Layers)-1 {
				prefix = "└── "
			}
			layerStr += fmt.Sprintf("%s%s\n", prefix, layer)
		}

		fmt.Printf("Image: %s\nId: %s\nParent: %s\nLayers: %d\n%sHistory: %s\n", imageID, inspect.ID, inspect.Parent, len(inspect.RootFS.Layers), layerStr, historyStr)

		fmt.Println("\n")

		if inspect.Parent == "" {
			break
		} else {
			imageID = inspect.Parent
		}
	}
}
