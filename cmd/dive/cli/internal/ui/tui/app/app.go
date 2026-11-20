package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/awesome-gocui/gocui"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"

	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui/key"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui/layout"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui/layout/compound"
)

const debug = false

// outputMode is swapped for gocui's simulated screen in tests
var outputMode = gocui.OutputNormal

// termSize is the size of the terminal gocui draws on (it opens /dev/tty too), an error means it can't be checked
var termSize = func() (int, int, error) {
	f, err := os.Open("/dev/tty")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	return term.GetSize(int(f.Fd()))
}

type app struct {
	gui        *gocui.Gui
	controller *controller
	layout     *layout.Manager
}

// Run is the UI entrypoint. It returns when the user quits or ctx is done.
func Run(ctx context.Context, c tui.Config) error {
	c.Preferences.Colors.Apply()
	var err error

	// there is nothing to show (the layer and tree views need a selected layer)
	if len(c.Analysis.Layers) == 0 {
		return errors.New("image has no layers")
	}

	// gocui waits for a resize on a terminal without a size, and with the terminal in raw mode ctrl-c can't stop it
	if outputMode == gocui.OutputNormal {
		if w, h, err := termSize(); err == nil && (w == 0 || h == 0) {
			return fmt.Errorf("the terminal has no size (%dx%d), set one (e.g. with `stty rows 40 cols 120`) and try again", w, h)
		}
	}

	// it appears there is a race condition where termbox.Init() will
	// block nearly indefinitely when running as the first process in
	// a Docker container when started within ~25ms of container startup.
	// I can't seem to determine the exact root cause, however, a large
	// enough sleep will prevent this behavior (todo: remove this hack)
	time.Sleep(100 * time.Millisecond)

	narrowAmbiguousWidth()

	g, err := gocui.NewGui(outputMode, true)
	if err != nil {
		return err
	}
	defer g.Close()

	_, err = newApp(g, c)
	if err != nil {
		return err
	}

	k, mod := gocui.MustParse("Ctrl+Z")
	if err := g.SetKeybinding("", k, mod, handle_ctrl_z); err != nil {
		return err
	}

	// a signal that isn't a key press (the terminal is in raw mode, ctrl-c is a key) cancels ctx
	stopped := make(chan struct{})
	defer close(stopped)
	go func() {
		select {
		case <-ctx.Done():
			g.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
		case <-stopped:
		}
	}()

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		return err
	}
	return nil
}

// narrowAmbiguousWidth counts east asian ambiguous width runes as one cell. go-runewidth counts them as two under a
// CJK locale, but terminals draw the tree and border box drawing characters (│ ├ └ ─) one cell wide, so every row
// would be laid out wrong. RUNEWIDTH_EASTASIAN still wins when it is set.
func narrowAmbiguousWidth() {
	if os.Getenv("RUNEWIDTH_EASTASIAN") == "" {
		runewidth.DefaultCondition.EastAsianWidth = false
	}
}

func newApp(gui *gocui.Gui, cfg tui.Config) (*app, error) {
	var err error
	var c *controller
	var globalHelpKeys []*key.Binding

	c, err = newController(gui, cfg)
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
