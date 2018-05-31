package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"github.com/jroimartin/gocui"
	"log"
)

var data struct {
	tree        *FileTree
	absPosition uint
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





func getAbsPositionNode() (node *FileNode) {
	var visiter func(*FileNode) error
	var evaluator func(*FileNode) bool
	var dfsCounter uint

	visiter = func(curNode *FileNode) error {
		if dfsCounter == data.absPosition {
			node = curNode
		}
		dfsCounter++
		return nil
	}

	evaluator = func(curNode *FileNode) bool {
		return !curNode.collapsed
	}

	err := data.tree.VisitDepthParentFirst(visiter, evaluator)
	if err != nil {
		// todo: you guessed it, check errors
	}

	return node
}

func showCurNodeInSideBar(g *gocui.Gui, v *gocui.View) error {
	g.Update(func(g *gocui.Gui) error {
		v, _ := g.View("side")
		// todo: handle above error.
		v.Clear()
		_, err := fmt.Fprintf(v, "FileNode:\n%+v\n", getAbsPositionNode())
		return err
	})
	// todo: blerg
	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()

		// if there isn't a next line
		line, err := v.Line(cy+1)
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
		data.absPosition++
		showCurNodeInSideBar(g, v)
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
		data.absPosition--
		showCurNodeInSideBar(g, v)
	}
	return nil
}


func toggleCollapse(g *gocui.Gui, v *gocui.View) error {
	node := getAbsPositionNode()
	node.collapsed = !node.collapsed
	return drawTree(g, v)
}

func drawTree(g *gocui.Gui, v *gocui.View) error {
	g.Update(func(g *gocui.Gui) error {
		v, _ := g.View("main")
		// todo: handle above error.
		v.Clear()
		_, err := fmt.Fprintln(v, data.tree.String())
		return err
	})
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
	if err := g.SetKeybinding("main", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeySpace, gocui.ModNone, toggleCollapse); err != nil {
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
		v.Wrap = true
	}
	if v, err := g.SetView("main", splitCol, -1, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = false
		v.Wrap = false
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		if _, err := g.SetCurrentView("main"); err != nil {
			return err
		}
		drawTree(g, v)
	}
	return nil
}


func main() {
	data.tree = NewTree()
	data.tree.AddPath("/etc/nginx/nginx.conf", nil)
	data.tree.AddPath("/etc/nginx/public", nil)
	data.tree.AddPath("/var/run/systemd", nil)
	data.tree.AddPath("/var/run/bashful", nil)
	data.tree.AddPath("/tmp", nil)
	data.tree.AddPath("/tmp/nonsense", nil)
	data.tree.AddPath("/tmp/wifi/coffeeyo", nil)

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


