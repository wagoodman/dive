package ui

import (
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/key"
	"github.com/wagoodman/dive/runtime/ui/layout"
	"github.com/wagoodman/dive/runtime/ui/layout/compound"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
)

const debug = false

// type global
type app struct {
	gui        *gocui.Gui
	controller *Controller
	layout     *layout.Manager
	detailedMode bool
}

var (
	once         sync.Once
	appSingleton *app
)

func newApp(gui *gocui.Gui, analysis *image.AnalysisResult, cache filetree.Comparer) (*app, error) {
	var err error
	once.Do(func() {
		var controller *Controller
		var globalHelpKeys []*key.Binding

		controller, err = NewController(gui, analysis, cache)
		if err != nil {
			return
		}

		// note: order matters when adding elements to the layout
		lm := layout.NewManager()
		lm.Add(controller.views.Status, layout.LocationFooter)
		lm.Add(controller.views.Filter, layout.LocationFooter)
		lm.Add(compound.NewLayerDetailsCompoundLayout(controller.views.Layer, controller.views.Details), layout.LocationColumn)
		lm.Add(controller.views.Tree, layout.LocationColumn)

		// todo: access this more programmatically
		if debug {
			lm.Add(controller.views.Debug, layout.LocationColumn)
		}
		gui.Cursor = false
		//g.Mouse = true
		gui.SetManagerFunc(lm.Layout)

		// var profileObj = profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook)
		//
		// onExit = func() {
		// 	profileObj.Stop()
		// }

		appSingleton = &app{
			gui:          gui,
			controller:   controller,
			layout:       lm,
			detailedMode: false,
		}

		var infos = []key.BindingInfo{
			{
				ConfigKeys: []string{"keybinding.quit"},
				OnAction:   appSingleton.quit,
				Display:    "Quit",
			},
			{
				ConfigKeys: []string{"keybinding.toggle-view"},
				OnAction:   controller.ToggleView,
				Display:    "Switch view",
			},
			{
				ConfigKeys: []string{"keybinding.filter-files"},
				OnAction:   controller.ToggleFilterView,
				IsSelected: controller.views.Filter.IsVisible,
				Display:    "Filter",
			},
		}

		globalHelpKeys, err = key.GenerateBindings(gui, "", infos)
		if err != nil {
			return
		}

		controller.views.Status.AddHelpKeys(globalHelpKeys...)

		// dont show these key bindings on the status pane
		quietKeys := []key.BindingInfo{
			{
				ConfigKeys: []string{"keybinding.toggle-details"},
				OnAction:   appSingleton.toggleDetails,
				Display:    "",
			},
		}
		_, err = key.GenerateBindings(gui, "", quietKeys)
		if err != nil {
			return
		}

		// perform the first update and render now that all resources have been loaded
		err = controller.UpdateAndRender()
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
// 	if controller.Tree != nil && controller.Tree.gui != nil {
// 		v, _ := controller.Tree.gui.View("debug")
// 		if v != nil {
// 			if len(v.BufferLines()) > 20 {
// 				v.Clear()
// 			}
// 			_, _ = fmt.Fprintln(v, s)
// 		}
// 	}
// }

// quit is the gocui callback invoked when the user hits Ctrl+C
func (a *app) quit() error {

	// profileObj.Stop()
	// onExit()

	return gocui.ErrQuit
}

func (a *app) toggleDetails() error {
	if a.detailedMode {
		err := a.layout.Remove(a.controller.views.Debug)
		if err != nil {
			logrus.Errorf("could not remove DEBUG pane")
		}
		a.layout.Add(a.controller.views.Tree, layout.LocationColumn)
		a.controller.views.Debug.ToggleHide()
		a.controller.views.Tree.ToggleHide()

		a.controller.views.Status.Retop()
		a.controller.views.Filter.Retop()

		a.detailedMode = false
	} else {
		err := a.layout.Remove(a.controller.views.Tree)
		if err != nil {
			logrus.Errorf("could not remove TREE pane")
		}
		a.layout.Add(a.controller.views.Debug, layout.LocationColumn)
		a.controller.views.Debug.ToggleHide()
		a.controller.views.Tree.ToggleHide()

		a.controller.views.Status.Retop()
		a.controller.views.Filter.Retop()

		a.detailedMode = true
	}

	return nil
}


// Run is the UI entrypoint.
func Run(analysis *image.AnalysisResult, treeStack filetree.Comparer) error {
	var err error

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer g.Close()

	_, err = newApp(g, analysis, treeStack)
	if err != nil {
		return err
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		logrus.Error("main loop error: ", err)
		return err
	}
	return nil
}
