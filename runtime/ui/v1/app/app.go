package app

import (
	"errors"
	"github.com/awesome-gocui/gocui"
	"github.com/sirupsen/logrus"
	v1 "github.com/wagoodman/dive/runtime/ui/v1"
	"github.com/wagoodman/dive/runtime/ui/v1/key"
	"github.com/wagoodman/dive/runtime/ui/v1/layout"
	"github.com/wagoodman/dive/runtime/ui/v1/layout/compound"
	"golang.org/x/net/context"
	"time"
)

const debug = false

type app struct {
	gui        *gocui.Gui
	controller *controller
	layout     *layout.Manager
}

// Run is the UI entrypoint.
func Run(ctx context.Context, c v1.Config) error {
	var err error

	// it appears there is a race condition where termbox.Init() will
	// block nearly indefinitely when running as the first process in
	// a Docker container when started within ~25ms of container startup.
	// I can't seem to determine the exact root cause, however, a large
	// enough sleep will prevent this behavior (todo: remove this hack)
	time.Sleep(100 * time.Millisecond)

	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		return err
	}
	defer g.Close()

	_, err = newApp(ctx, g, c)
	if err != nil {
		return err
	}

	k, mod := gocui.MustParse("Ctrl+Z")
	if err := g.SetKeybinding("", k, mod, handle_ctrl_z); err != nil {
		return err
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		logrus.Error("main loop error: ", err)
		return err
	}
	return nil
}

func newApp(ctx context.Context, gui *gocui.Gui, cfg v1.Config) (*app, error) {
	var err error
	var c *controller
	var globalHelpKeys []*key.Binding

	c, err = newController(ctx, gui, cfg)
	if err != nil {
		return nil, err
	}

	// note: order matters when adding elements to the layout
	lm := layout.NewManager()
	lm.Add(c.views.Status, layout.LocationFooter)
	lm.Add(c.views.Filter, layout.LocationFooter)
	lm.Add(compound.NewLayerDetailsCompoundLayout(c.views.Layer, c.views.LayerDetails, c.views.ImageDetails), layout.LocationColumn)
	lm.Add(c.views.Tree, layout.LocationColumn)

	// todo: access this more programmatically
	if debug {
		lm.Add(c.views.Debug, layout.LocationColumn)
	}
	gui.Cursor = false
	// g.Mouse = true
	gui.SetManagerFunc(lm.Layout)

	a := &app{
		gui:        gui,
		controller: c,
		layout:     lm,
	}

	var infos = []key.BindingInfo{
		{
			Config:   cfg.Preferences.KeyBindings.Global.Quit,
			OnAction: a.quit,
			Display:  "Quit",
		},
		{
			Config:   cfg.Preferences.KeyBindings.Global.ToggleView,
			OnAction: c.ToggleView,
			Display:  "Switch view",
		},
		{
			Config:   cfg.Preferences.KeyBindings.Navigation.Right,
			OnAction: c.NextPane,
		},
		{
			Config:   cfg.Preferences.KeyBindings.Navigation.Left,
			OnAction: c.PrevPane,
		},
		{
			Config:     cfg.Preferences.KeyBindings.Global.FilterFiles,
			OnAction:   c.ToggleFilterView,
			IsSelected: c.views.Filter.IsVisible,
			Display:    "Filter",
		},
		{
			Config:   cfg.Preferences.KeyBindings.Global.CloseFilterFiles,
			OnAction: c.CloseFilterView,
		},
	}

	globalHelpKeys, err = key.GenerateBindings(gui, "", infos)
	if err != nil {
		return nil, err
	}

	c.views.Status.AddHelpKeys(globalHelpKeys...)

	// perform the first update and render now that all resources have been loaded
	err = c.UpdateAndRender()
	if err != nil {
		return nil, err
	}

	return a, err
}

// quit is the gocui callback invoked when the user hits Ctrl+C
func (a *app) quit() error {
	return gocui.ErrQuit
}
