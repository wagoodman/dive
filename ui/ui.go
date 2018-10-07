package ui

import (
	"errors"
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
	"github.com/wagoodman/dive/filetree"
	"github.com/wagoodman/dive/image"
	"github.com/fatih/color"
	"os"
	"runtime"
	"runtime/pprof"
)

const debug = false
const profile = false

func debugPrint(s string) {
	if debug && Views.Tree != nil && Views.Tree.gui != nil {
		v, _ := Views.Tree.gui.View("debug")
		if v != nil {
			if len(v.BufferLines()) > 20 {
				v.Clear()
			}
			_, _ = fmt.Fprintln(v, s)
		}
	}
}

var Formatting struct {
	Header        func(...interface{})(string)
	Selected      func(...interface{})(string)
	StatusSelected      func(...interface{})(string)
	StatusNormal      func(...interface{})(string)
	StatusControlSelected      func(...interface{})(string)
	StatusControlNormal      func(...interface{})(string)
	CompareTop    func(...interface{})(string)
	CompareBottom func(...interface{})(string)
}

var Views struct {
	Tree   *FileTreeView
	Layer  *LayerView
	Status *StatusView
	Filter *FilterView
	lookup map[string]View
}

type View interface {
	Setup(*gocui.View, *gocui.View) error
	CursorDown() error
	CursorUp() error
	Render() error
	Update() error
	KeyHelp() string
	IsVisible() bool
}

func toggleView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == Views.Layer.Name {
		_, err := g.SetCurrentView(Views.Tree.Name)
		Update()
		Render()
		return err
	}
	_, err := g.SetCurrentView(Views.Layer.Name)
	Update()
	Render()
	return err
}

func toggleFilterView(g *gocui.Gui, v *gocui.View) error {
	// delete all user input from the tree view
	Views.Filter.view.Clear()
	Views.Filter.view.SetCursor(0,0)

	// toggle hiding
	Views.Filter.hidden = !Views.Filter.hidden

	if !Views.Filter.hidden {
		_, err := g.SetCurrentView(Views.Filter.Name)
		if err != nil {
			return err
		}
		Update()
		Render()
	} else {
		toggleView(g, v)
	}

	return nil
}

func CursorDown(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()

	// if there isn't a next line
	line, err := v.Line(cy + 1)
	if err != nil {
		// todo: handle error
	}
	if len(line) == 0 {
		return errors.New("unable to move cursor down, empty line")
	}
	if err := v.SetCursor(cx, cy+1); err != nil {
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+1); err != nil {
			return err
		}
	}
	return nil
}

func CursorUp(g *gocui.Gui, v *gocui.View) error {
	ox, oy := v.Origin()
	cx, cy := v.Cursor()
	if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
		if err := v.SetOrigin(ox, oy-1); err != nil {
			return err
		}
	}
	return nil
}


var cpuProfilePath *os.File
var memoryProfilePath *os.File

func quit(g *gocui.Gui, v *gocui.View) error {
	if profile {
		pprof.StopCPUProfile()
		runtime.GC() // get up-to-date statistics
		pprof.WriteHeapProfile(memoryProfilePath)
		memoryProfilePath.Close()
		cpuProfilePath.Close()
	}
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	//if err := g.SetKeybinding("main", gocui.MouseLeft, gocui.ModNone, toggleCollapse); err != nil {
	//	return err
	//}
	if err := g.SetKeybinding("", gocui.KeyCtrlSpace, gocui.ModNone, toggleView); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlSlash, gocui.ModNone, toggleFilterView); err != nil {
		return err
	}

	return nil
}

func isNewView(errs ...error) bool {
	for _, err := range errs {
		if err == nil {
			return false
		}
		if err != nil && err != gocui.ErrUnknownView {
			return false
		}
	}
	return true
}

// TODO: this logic should be refactored into an abstraction that takes care of the math for us
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	splitCols := maxX / 2
	debugWidth := 0
	if debug {
		debugWidth = maxX / 4
	}
	debugCols := maxX - debugWidth
	bottomRows := 1
	headerRows := 1

	filterBarHeight := 1
	statusBarHeight := 1

	statusBarIndex := 1
	filterBarIndex := 2

	var view, header *gocui.View
	var viewErr, headerErr, err error

	if Views.Filter.hidden {
		bottomRows--
		filterBarHeight = 0
	}

	// Debug pane
	if debug {
		if _, err := g.SetView("debug", debugCols, -1, maxX, maxY-bottomRows); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
		}
	}

	// Layers
	view, viewErr = g.SetView(Views.Layer.Name, -1, -1+headerRows, splitCols, maxY-bottomRows)
	header, headerErr = g.SetView(Views.Layer.Name+"header", -1, -1, splitCols, headerRows)
	if isNewView(viewErr, headerErr) {
		Views.Layer.Setup(view, header)

		if _, err = g.SetCurrentView(Views.Layer.Name); err != nil {
			return err
		}
	}

	// Filetree
	view, viewErr = g.SetView(Views.Tree.Name, splitCols, -1+headerRows, debugCols, maxY-bottomRows)
	header, headerErr = g.SetView(Views.Tree.Name+"header", splitCols, -1, debugCols, headerRows)
	if isNewView(viewErr, headerErr) {
		Views.Tree.Setup(view, header)
	}

	// Status Bar
	view, viewErr = g.SetView(Views.Status.Name, -1, maxY-statusBarHeight-statusBarIndex, maxX, maxY-(statusBarIndex-1))
	if isNewView(viewErr, headerErr) {
		Views.Status.Setup(view, nil)
	}

	// Filter Bar
	view, viewErr = g.SetView(Views.Filter.Name, len(Views.Filter.headerStr)-1, maxY-filterBarHeight-filterBarIndex, maxX, maxY-(filterBarIndex-1))
	header, headerErr = g.SetView(Views.Filter.Name+"header", -1, maxY-filterBarHeight - filterBarIndex, len(Views.Filter.headerStr), maxY-(filterBarIndex-1))
	if isNewView(viewErr, headerErr) {
		Views.Filter.Setup(view, header)
	}


	return nil
}

func Update() {
	for _, view := range Views.lookup {
		view.Update()
	}
}

func Render() {
	for _, view := range Views.lookup {
		if view.IsVisible() {
			view.Render()
		}
	}
}

func renderStatusOption(control, title string, selected bool) string {
	if selected {
		return Formatting.StatusSelected("▏") + Formatting.StatusControlSelected(control) +  Formatting.StatusSelected("  " + title + " ")
	} else {
		return Formatting.StatusNormal("▏") + Formatting.StatusControlNormal(control) +  Formatting.StatusNormal("  " + title + " ")
	}
}

func Run(layers []*image.Layer, refTrees []*filetree.FileTree) {

	Formatting.Selected = color.New(color.ReverseVideo, color.Bold).SprintFunc()
	Formatting.Header = color.New(color.Bold).SprintFunc()
	Formatting.StatusSelected = color.New(color.BgMagenta, color.FgWhite).SprintFunc()
	Formatting.StatusNormal = color.New(color.ReverseVideo).SprintFunc()
	Formatting.StatusControlSelected = color.New(color.BgMagenta, color.FgWhite, color.Bold).SprintFunc()
	Formatting.StatusControlNormal = color.New(color.ReverseVideo, color.Bold).SprintFunc()
	Formatting.CompareTop = color.New(color.BgMagenta).SprintFunc()
	Formatting.CompareBottom = color.New(color.BgGreen).SprintFunc()

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	Views.lookup = make(map[string]View)

	Views.Layer = NewLayerView("side", g, layers)
	Views.lookup[Views.Layer.Name] = Views.Layer

	Views.Tree = NewFileTreeView("main", g, filetree.StackRange(refTrees, 0,0), refTrees)
	Views.lookup[Views.Tree.Name] = Views.Tree

	Views.Status = NewStatusView("status", g)
	Views.lookup[Views.Status.Name] = Views.Status

	Views.Filter = NewFilterView("command", g)
	Views.lookup[Views.Filter.Name] = Views.Filter

	g.Cursor = false
	//g.Mouse = true
	g.SetManagerFunc(layout)

	// let the default position of the cursor be the last layer
	// Views.Layer.SetCursor(len(Views.Layer.Layers)-1)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if profile {
		os.Create("cpu.pprof")
		os.Create("mem.pprof")
		pprof.StartCPUProfile(cpuProfilePath)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
