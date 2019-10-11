package ui

import (
	"errors"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/key"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/dive/filetree"
)

const debug = false

// type global
type app struct {
	gui *gocui.Gui
	controllers *controllerCollection
}

var (
	once         sync.Once
	appSingleton *app
)

func newApp(gui *gocui.Gui, analysis *image.AnalysisResult, cache filetree.TreeCache) (*app, error) {
	var err error
	once.Do(func() {
		var theControls *controllerCollection
		var globalHelpKeys []*key.Binding

		theControls, err = newControllerCollection(gui, analysis, cache)
		if err != nil {
			return
		}

		gui.Cursor = false
		//g.Mouse = true
		gui.SetManagerFunc(layout)

		// var profileObj = profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook)
		//
		// onExit = func() {
		// 	profileObj.Stop()
		// }


		appSingleton = &app{
			gui:         gui,
			controllers: theControls,
		}

		var infos = []key.BindingInfo{
			{
				ConfigKeys: []string{"keybinding.quit"},
				OnAction:   quit,
				Display:    "Quit",
			},
			{
				ConfigKeys: []string{"keybinding.toggle-view"},
				OnAction:   appSingleton.toggleView,
				Display:    "Switch view",
			},
			{
				ConfigKeys: []string{"keybinding.filter-files"},
				OnAction:   appSingleton.toggleFilterView,
				IsSelected: controllers.Filter.IsVisible,
				Display:    "Filter",
			},
		}

		globalHelpKeys, err = key.GenerateBindings(gui, "", infos)
		if err != nil {
			return
		}

		theControls.Status.AddHelpKeys(globalHelpKeys...)


		// perform the first update and render now that all resources have been loaded
		err = UpdateAndRender()
		if err != nil {
			return
		}


	})

	return appSingleton, err
}

// var profileObj = profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook)
// var onExit func()

// debugPrint writes the given string to the debug pane (if the debug pane is enabled)
// func debugPrint(s string) {
// 	if controllers.Tree != nil && controllers.Tree.gui != nil {
// 		v, _ := controllers.Tree.gui.View("debug")
// 		if v != nil {
// 			if len(v.BufferLines()) > 20 {
// 				v.Clear()
// 			}
// 			_, _ = fmt.Fprintln(v, s)
// 		}
// 	}
// }

var lastX, lastY int

func UpdateAndRender() error {
	err := Update()
	if err != nil {
		logrus.Debug("failed update: ", err)
		return err
	}

	err = Render()
	if err != nil {
		logrus.Debug("failed render: ", err)
		return err
	}

	return nil
}

// toggleView switches between the file view and the layer view and re-renders the screen.
func (ui *app) toggleView() (err error) {
	v := ui.gui.CurrentView()
	if v == nil || v.Name() == controllers.Layer.name {
		_, err = ui.gui.SetCurrentView(controllers.Tree.name)
	} else {
		_, err = ui.gui.SetCurrentView(controllers.Layer.name)
	}

	if err != nil {
		logrus.Error("unable to toggle view: ", err)
		return err
	}

	return UpdateAndRender()
}

// toggleFilterView shows/hides the file tree filter pane.
func (ui *app) toggleFilterView() error {
	// delete all user input from the tree view
	controllers.Filter.view.Clear()

	// toggle hiding
	controllers.Filter.hidden = !controllers.Filter.hidden

	if !controllers.Filter.hidden {
		_, err := ui.gui.SetCurrentView(controllers.Filter.name)
		if err != nil {
			logrus.Error("unable to toggle filter view: ", err)
			return err
		}
		return UpdateAndRender()
	}

	err := ui.toggleView()
	if err != nil {
		logrus.Error("unable to toggle filter view (back): ", err)
		return err
	}

	err = controllers.Filter.view.SetCursor(0, 0)
	if err != nil {
		return err
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
		return err
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
func quit() error {

	// profileObj.Stop()
	// onExit()

	return gocui.ErrQuit
}

// isNewView determines if a view has already been created based on the set of errors given (a bit hokie)
func isNewView(errs ...error) bool {
	for _, err := range errs {
		if err == nil {
			return false
		}
		if err != gocui.ErrUnknownView {
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

	layersHeight := len(controllers.Layer.Layers) + headerRows + 1 // layers + header + base image layer row
	maxLayerHeight := int(0.75 * float64(maxY))
	if layersHeight > maxLayerHeight {
		layersHeight = maxLayerHeight
	}

	var view, header *gocui.View
	var viewErr, headerErr, err error

	if controllers.Filter.hidden {
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
	view, viewErr = g.SetView(controllers.Layer.name, -1, -1+headerRows, splitCols, layersHeight)
	header, headerErr = g.SetView(controllers.Layer.name+"header", -1, -1, splitCols, headerRows)
	if isNewView(viewErr, headerErr) {
		err = controllers.Layer.Setup(view, header)
		if err != nil {
			logrus.Error("unable to setup layer controller", err)
			return err
		}

		if _, err = g.SetCurrentView(controllers.Layer.name); err != nil {
			logrus.Error("unable to set view to layer", err)
			return err
		}
		// since we are selecting the view, we should rerender to indicate it is selected
		err = controllers.Layer.Render()
		if err != nil {
			logrus.Error("unable to render layer view", err)
			return err
		}
	}

	// Details
	view, viewErr = g.SetView(controllers.Details.name, -1, -1+layersHeight+headerRows, splitCols, maxY-bottomRows)
	header, headerErr = g.SetView(controllers.Details.name+"header", -1, -1+layersHeight, splitCols, layersHeight+headerRows)
	if isNewView(viewErr, headerErr) {
		err = controllers.Details.Setup(view, header)
		if err != nil {
			return err
		}
	}

	// Filetree
	offset := 0
	if !controllers.Tree.vm.ShowAttributes {
		offset = 1
	}
	view, viewErr = g.SetView(controllers.Tree.name, splitCols, -1+headerRows-offset, debugCols, maxY-bottomRows)
	header, headerErr = g.SetView(controllers.Tree.name+"header", splitCols, -1, debugCols, headerRows-offset)
	if isNewView(viewErr, headerErr) {
		err = controllers.Tree.Setup(view, header)
		if err != nil {
			logrus.Error("unable to setup tree controller", err)
			return err
		}
	}
	err = controllers.Tree.onLayoutChange(resized)
	if err != nil {
		logrus.Error("unable to setup layer controller onLayoutChange", err)
		return err
	}

	// Status Bar
	view, viewErr = g.SetView(controllers.Status.name, -1, maxY-statusBarHeight-statusBarIndex, maxX, maxY-(statusBarIndex-1))
	if isNewView(viewErr, headerErr) {
		err = controllers.Status.Setup(view, nil)
		if err != nil {
			logrus.Error("unable to setup status controller", err)
			return err
		}
	}

	// Filter Bar
	view, viewErr = g.SetView(controllers.Filter.name, len(controllers.Filter.headerStr)-1, maxY-filterBarHeight-filterBarIndex, maxX, maxY-(filterBarIndex-1))
	header, headerErr = g.SetView(controllers.Filter.name+"header", -1, maxY-filterBarHeight-filterBarIndex, len(controllers.Filter.headerStr), maxY-(filterBarIndex-1))
	if isNewView(viewErr, headerErr) {
		err = controllers.Filter.Setup(view, header)
		if err != nil {
			logrus.Error("unable to setup filter controller", err)
			return err
		}
	}

	return nil
}

// Update refreshes the state objects for future rendering.
func Update() error {
	for _, controller := range controllers.lookup {
		err := controller.Update()
		if err != nil {
			logrus.Debug("unable to update controller: ")
			return err
		}
	}
	return nil
}

// Render flushes the state objects to the screen.
func Render() error {
	for _, controller := range controllers.lookup {
		if controller.IsVisible() {
			err := controller.Render()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Run is the UI entrypoint.
func Run(analysis *image.AnalysisResult, cache filetree.TreeCache) error {
	var err error

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer g.Close()

	_, err = newApp(g, analysis, cache)
	if err != nil {
		return err
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		logrus.Error("main loop error: ", err)
		return err
	}
	return nil
}
