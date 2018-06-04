package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"log"

	"github.com/docker/docker/client"
	"github.com/jroimartin/gocui"
	"golang.org/x/net/context"
)

var data struct {
	refTrees []*FileTree
	manifest *Manifest
}

var views struct {
	treeView  *FileTreeView
	layerView *LayerView
}

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

func demo() {
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

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func nextView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == views.layerView.name {
		_, err := g.SetCurrentView(views.treeView.name)
		return err
	}
	_, err := g.SetCurrentView(views.layerView.name)
	return err
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()

	// if there isn't a next line
	line, err := v.Line(cy + 1)
	if err != nil {
		// todo: handle error
	}
	if len(line) == 0 {
		return nil
	}
	if err := v.SetCursor(cx, cy+1); err != nil {
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+1); err != nil {
			return err
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	ox, oy := v.Origin()
	cx, cy := v.Cursor()
	if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
		if err := v.SetOrigin(ox, oy-1); err != nil {
			return err
		}
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	//if err := g.SetKeybinding("main", gocui.MouseLeft, gocui.ModNone, toggleCollapse); err != nil {
	//	return err
	//}
	if err := g.SetKeybinding("side", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}

	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	splitCol := 50
	if v, err := g.SetView("side", -1, -1, splitCol, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		views.layerView = NewLayerView("side", g, v, data.manifest)
		views.layerView.keybindings()

	}
	if v, err := g.SetView("main", splitCol, -1, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		views.treeView = NewFileTreeView("main", g, v, StackRange(data.refTrees, 0))
		views.treeView.keybindings()

		if _, err := g.SetCurrentView(views.treeView.name); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	demo()
	initializeData()

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = false
	//g.Mouse = true
	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

}
