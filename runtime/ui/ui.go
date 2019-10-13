package ui

import (
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/controller"
	"github.com/wagoodman/dive/runtime/ui/key"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
)

const debug = false

// type global
type app struct {
	gui         *gocui.Gui
	controllers *controller.ControllerCollection
	layout      *layoutManager
}

var (
	once         sync.Once
	appSingleton *app
)

func newApp(gui *gocui.Gui, analysis *image.AnalysisResult, cache filetree.TreeCache) (*app, error) {
	var err error
	once.Do(func() {
		var theControls *controller.ControllerCollection
		var globalHelpKeys []*key.Binding

		theControls, err = controller.NewControllerCollection(gui, analysis, cache)
		if err != nil {
			return
		}

		lm := newLayoutManager(theControls)

		gui.Cursor = false
		//g.Mouse = true
		gui.SetManagerFunc(lm.layout)

		// var profileObj = profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook)
		//
		// onExit = func() {
		// 	profileObj.Stop()
		// }

		appSingleton = &app{
			gui:         gui,
			controllers: theControls,
			layout:      lm,
		}

		var infos = []key.BindingInfo{
			{
				ConfigKeys: []string{"keybinding.quit"},
				OnAction:   appSingleton.quit,
				Display:    "Quit",
			},
			{
				ConfigKeys: []string{"keybinding.toggle-view"},
				OnAction:   theControls.ToggleView,
				Display:    "Switch view",
			},
			{
				ConfigKeys: []string{"keybinding.filter-files"},
				OnAction:   theControls.ToggleFilterView,
				IsSelected: theControls.Filter.IsVisible,
				Display:    "Filter",
			},
		}

		globalHelpKeys, err = key.GenerateBindings(gui, "", infos)
		if err != nil {
			return
		}

		theControls.Status.AddHelpKeys(globalHelpKeys...)

		// perform the first update and render now that all resources have been loaded
		err = theControls.UpdateAndRender()
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

// quit is the gocui callback invoked when the user hits Ctrl+C
func (ui *app) quit() error {

	// profileObj.Stop()
	// onExit()

	return gocui.ErrQuit
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
