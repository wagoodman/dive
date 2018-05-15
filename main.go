package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func saveImage(readCloser io.ReadCloser) {
	defer readCloser.Close()

	path := "image"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}

	fo, err := os.Create("image/cache.tar")
	check(err)

	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()
	w := bufio.NewWriter(fo)

	buf := make([]byte, 1024)
	for {
		n, err := readCloser.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		if _, err := w.Write(buf[:n]); err != nil {
			panic(err)
		}
	}

	if err = w.Flush(); err != nil {
		panic(err)
	}
}

func main() {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	// imageID := "golang:alpine"
	imageID := "die-test:latest"

	fmt.Println("Saving Image...")
	readCloser, err := cli.ImageSave(ctx, []string{imageID})
	check(err)
	saveImage(readCloser)

	for {
		inspect, _, err := cli.ImageInspectWithRaw(ctx, imageID)
		check(err)

		history, err := cli.ImageHistory(ctx, imageID)
		check(err)

		historyStr, err := json.MarshalIndent(history, "", "  ")
		check(err)

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
	fmt.Println("See './image' for the cached image tar")
}
