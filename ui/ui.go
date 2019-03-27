package ui

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/filetree"
	"github.com/wagoodman/dive/image"
	"github.com/wagoodman/dive/utils"
	"github.com/wagoodman/keybinding"
)

const debug = false

// var profileObj = profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook)
// var onExit func()

// debugPrint writes the given string to the debug pane (if the debug pane is enabled)
func debugPrint(s string) {
	if debug && Controllers.Tree != nil && Controllers.Tree.gui != nil {
		v, _ := Controllers.Tree.gui.View("debug")
		if v != nil {
			if len(v.BufferLines()) > 20 {
				v.Clear()
			}
			_, _ = fmt.Fprintln(v, s)
		}
	}
}

// Formatting defines standard functions for formatting UI sections.
var Formatting struct {
	Header                func(...interface{}) string
	Selected              func(...interface{}) string
	StatusSelected        func(...interface{}) string
	StatusNormal          func(...interface{}) string
	StatusControlSelected func(...interface{}) string
	StatusControlNormal   func(...interface{}) string
	CompareTop            func(...interface{}) string
	CompareBottom         func(...interface{}) string
}

// Controllers contains all rendered UI panes.
var Controllers struct {
	Tree    *FileTreeController
	Layer   *LayerController
	Status  *StatusController
	Filter  *FilterController
	Details *DetailsController
	lookup  map[string]View
}

var GlobalKeybindings struct {
	quit       []keybinding.Key
	toggleView []keybinding.Key
	filterView []keybinding.Key
}

var lastX, lastY int

// View defines the a renderable terminal screen pane.
type View interface {
	Setup(*gocui.View, *gocui.View) error
	CursorDown() error
	CursorUp() error
	Render() error
	Update() error
	KeyHelp() string
	IsVisible() bool
}

// toggleView switches between the file view and the layer view and re-renders the screen.
func toggleView(g *gocui.Gui, v *gocui.View) (err error) {
	if v == nil || v.Name() == Controllers.Layer.Name {
		_, err = g.SetCurrentView(Controllers.Tree.Name)
	} else {
		_, err = g.SetCurrentView(Controllers.Layer.Name)
	}
	Update()
	Render()
	return err
}

// toggleFilterView shows/hides the file tree filter pane.
func toggleFilterView(g *gocui.Gui, v *gocui.View) error {
	// delete all user input from the tree view
	Controllers.Filter.view.Clear()
	Controllers.Filter.view.SetCursor(0, 0)

	// toggle hiding
	Controllers.Filter.hidden = !Controllers.Filter.hidden

	if !Controllers.Filter.hidden {
		_, err := g.SetCurrentView(Controllers.Filter.Name)
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

// CursorDown moves the cursor down in the currently selected gocui pane, scrolling the screen as needed.
func CursorDown(g *gocui.Gui, v *gocui.View) error {
	return CursorStep(g, v, 1)
}

// CursorUp moves the cursor up in the currently selected gocui pane, scrolling the screen as needed.
func CursorUp(g *gocui.Gui, v *gocui.View) error {
	return CursorStep(g, v, -1)
}

// Moves the cursor the given step distance, setting the origin to the new cursor line
func CursorStep(g *gocui.Gui, v *gocui.View, step int) error {
	cx, cy := v.Cursor()

	// if there isn't a next line
	line, err := v.Line(cy + step)
	if err != nil {
		// todo: handle error
	}
	if len(line) == 0 {
		return errors.New("unable to move the cursor, empty line")
	}
	if err := v.SetCursor(cx, cy+step); err != nil {
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+step); err != nil {
			return err
		}
	}
	return nil
}

// quit is the gocui callback invoked when the user hits Ctrl+C
func quit(g *gocui.Gui, v *gocui.View) error {

	// profileObj.Stop()
	// onExit()

	return gocui.ErrQuit
}

// keyBindings registers global key press actions, valid when in any pane.
func keyBindings(g *gocui.Gui) error {
	for _, key := range GlobalKeybindings.quit {
		if err := g.SetKeybinding("", key.Value, key.Modifier, quit); err != nil {
			return err
		}
	}

	for _, key := range GlobalKeybindings.toggleView {
		if err := g.SetKeybinding("", key.Value, key.Modifier, toggleView); err != nil {
			return err
		}
	}

	for _, key := range GlobalKeybindings.filterView {
		if err := g.SetKeybinding("", key.Value, key.Modifier, toggleFilterView); err != nil {
			return err
		}
	}

	return nil
}

// isNewView determines if a view has already been created based on the set of errors given (a bit hokie)
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

// layout defines the definition of the window pane size and placement relations to one another. This
// is invoked at application start and whenever the screen dimensions change.
func layout(g *gocui.Gui) error {
	// TODO: this logic should be refactored into an abstraction that takes care of the math for us

	maxX, maxY := g.Size()
	var resized bool
	if maxX != lastX {
		resized = true
	}
	if maxY != lastY {
		resized = true
	}
	lastX, lastY = maxX, maxY
	fileTreeSplitRatio := viper.GetFloat64("filetree.pane-width")
	if fileTreeSplitRatio >= 1 || fileTreeSplitRatio <= 0 {
		logrus.Errorf("invalid config value: 'filetree.pane-width' should be 0 < value < 1, given '%v'", fileTreeSplitRatio)
		fileTreeSplitRatio = 0.5
	}
	splitCols := int(float64(maxX) * (1.0 - fileTreeSplitRatio))
	debugWidth := 0
	if debug {
		debugWidth = maxX / 4
	}
	debugCols := maxX - debugWidth
	bottomRows := 1
	headerRows := 2

	filterBarHeight := 1
	statusBarHeight := 1

	statusBarIndex := 1
	filterBarIndex := 2

	layersHeight := len(Controllers.Layer.Layers) + headerRows + 1 // layers + header + base image layer row
	maxLayerHeight := int(0.75 * float64(maxY))
	if layersHeight > maxLayerHeight {
		layersHeight = maxLayerHeight
	}

	var view, header *gocui.View
	var viewErr, headerErr, err error

	if Controllers.Filter.hidden {
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
	view, viewErr = g.SetView(Controllers.Layer.Name, -1, -1+headerRows, splitCols, layersHeight)
	header, headerErr = g.SetView(Controllers.Layer.Name+"header", -1, -1, splitCols, headerRows)
	if isNewView(viewErr, headerErr) {
		Controllers.Layer.Setup(view, header)

		if _, err = g.SetCurrentView(Controllers.Layer.Name); err != nil {
			return err
		}
		// since we are selecting the view, we should rerender to indicate it is selected
		Controllers.Layer.Render()
	}

	// Details
	view, viewErr = g.SetView(Controllers.Details.Name, -1, -1+layersHeight+headerRows, splitCols, maxY-bottomRows)
	header, headerErr = g.SetView(Controllers.Details.Name+"header", -1, -1+layersHeight, splitCols, layersHeight+headerRows)
	if isNewView(viewErr, headerErr) {
		Controllers.Details.Setup(view, header)
	}

	// Filetree
	offset := 0
	if !Controllers.Tree.vm.ShowAttributes {
		offset = 1
	}
	view, viewErr = g.SetView(Controllers.Tree.Name, splitCols, -1+headerRows-offset, debugCols, maxY-bottomRows)
	header, headerErr = g.SetView(Controllers.Tree.Name+"header", splitCols, -1, debugCols, headerRows-offset)
	if isNewView(viewErr, headerErr) {
		Controllers.Tree.Setup(view, header)
	}
	Controllers.Tree.onLayoutChange(resized)

	// Status Bar
	view, viewErr = g.SetView(Controllers.Status.Name, -1, maxY-statusBarHeight-statusBarIndex, maxX, maxY-(statusBarIndex-1))
	if isNewView(viewErr, headerErr) {
		Controllers.Status.Setup(view, nil)
	}

	// Filter Bar
	view, viewErr = g.SetView(Controllers.Filter.Name, len(Controllers.Filter.headerStr)-1, maxY-filterBarHeight-filterBarIndex, maxX, maxY-(filterBarIndex-1))
	header, headerErr = g.SetView(Controllers.Filter.Name+"header", -1, maxY-filterBarHeight-filterBarIndex, len(Controllers.Filter.headerStr), maxY-(filterBarIndex-1))
	if isNewView(viewErr, headerErr) {
		Controllers.Filter.Setup(view, header)
	}

	return nil
}

// Update refreshes the state objects for future rendering.
func Update() {
	for _, view := range Controllers.lookup {
		view.Update()
	}
}

// Render flushes the state objects to the screen.
func Render() {
	for _, view := range Controllers.lookup {
		if view.IsVisible() {
			view.Render()
		}
	}
}

// renderStatusOption formats key help bindings-to-title pairs.
func renderStatusOption(control, title string, selected bool) string {
	if selected {
		return Formatting.StatusSelected("▏") + Formatting.StatusControlSelected(control) + Formatting.StatusSelected(" "+title+" ")
	} else {
		return Formatting.StatusNormal("▏") + Formatting.StatusControlNormal(control) + Formatting.StatusNormal(" "+title+" ")
	}
}

// Run is the UI entrypoint.
func Run(analysis *image.AnalysisResult, cache filetree.TreeCache) {

	Formatting.Selected = color.New(color.ReverseVideo, color.Bold).SprintFunc()
	Formatting.Header = color.New(color.Bold).SprintFunc()
	Formatting.StatusSelected = color.New(color.BgMagenta, color.FgWhite).SprintFunc()
	Formatting.StatusNormal = color.New(color.ReverseVideo).SprintFunc()
	Formatting.StatusControlSelected = color.New(color.BgMagenta, color.FgWhite, color.Bold).SprintFunc()
	Formatting.StatusControlNormal = color.New(color.ReverseVideo, color.Bold).SprintFunc()
	Formatting.CompareTop = color.New(color.BgMagenta).SprintFunc()
	Formatting.CompareBottom = color.New(color.BgGreen).SprintFunc()

	var err error
	GlobalKeybindings.quit, err = keybinding.ParseAll(viper.GetString("keybinding.quit"))
	if err != nil {
		logrus.Error(err)
	}
	GlobalKeybindings.toggleView, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-view"))
	if err != nil {
		logrus.Error(err)
	}
	GlobalKeybindings.filterView, err = keybinding.ParseAll(viper.GetString("keybinding.filter-files"))
	if err != nil {
		logrus.Error(err)
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		logrus.Error(err)
	}
	utils.SetUi(g)
	defer g.Close()

	Controllers.lookup = make(map[string]View)

	Controllers.Layer = NewLayerController("side", g, analysis.Layers)
	Controllers.lookup[Controllers.Layer.Name] = Controllers.Layer

	Controllers.Tree = NewFileTreeController("main", g, filetree.StackTreeRange(analysis.RefTrees, 0, 0), analysis.RefTrees, cache)
	Controllers.lookup[Controllers.Tree.Name] = Controllers.Tree

	Controllers.Status = NewStatusController("status", g)
	Controllers.lookup[Controllers.Status.Name] = Controllers.Status

	Controllers.Filter = NewFilterController("command", g)
	Controllers.lookup[Controllers.Filter.Name] = Controllers.Filter

	Controllers.Details = NewDetailsController("details", g, analysis.Efficiency, analysis.Inefficiencies)
	Controllers.lookup[Controllers.Details.Name] = Controllers.Details

	g.Cursor = false
	//g.Mouse = true
	g.SetManagerFunc(layout)

	// var profileObj = profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	//
	// onExit = func() {
	// 	profileObj.Stop()
	// }

	// perform the first update and render now that all resources have been loaded
	Update()
	Render()

	if err := keyBindings(g); err != nil {
		logrus.Error(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		logrus.Error(err)
	}
	utils.Exit(0)
}
