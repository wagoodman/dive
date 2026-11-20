package app

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/awesome-gocui/gocui"

	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui/view"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui/viewmodel"
	"github.com/wagoodman/dive/internal/bus"
)

type controller struct {
	gui    *gocui.Gui
	views  *view.Views
	config tui.Config

	// the focused view (and its key help) from before the filter was opened, focus goes back to it on close
	preFilterView string
	preFilterHelp view.Helper
}

func newController(g *gocui.Gui, cfg tui.Config) (*controller, error) {
	cfg.Preferences.Colors.Apply()
	views, err := view.NewViews(g, cfg)
	if err != nil {
		return nil, err
	}

	c := &controller{
		gui:    g,
		views:  views,
		config: cfg,
	}

	// layer view cursor down event should trigger an update in the file tree
	c.views.Layer.AddLayerChangeListener(c.onLayerChange)

	// update the status pane when a filetree option is changed by the user
	c.views.Tree.AddViewOptionChangeListener(c.onFileTreeViewOptionChange)

	// update the status pane when a filetree option is changed by the user
	c.views.Tree.AddViewExtractListener(c.onFileTreeViewExtract)

	// update the tree view while the user types into the filter view
	c.views.Filter.AddFilterEditListener(c.onFilterEdit)

	// propagate initial conditions to necessary views
	err = c.onLayerChange(viewmodel.LayerSelection{
		Layer:           c.views.Layer.CurrentLayer(),
		BottomTreeStart: 0,
		BottomTreeStop:  0,
		TopTreeStart:    0,
		TopTreeStop:     0,
	})

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *controller) onFileTreeViewExtract(p string) error {
	// files land relative to the current working directory, the same as before
	dst, err := filepath.Abs(".")
	if err == nil {
		err = extract(c.config.Analysis.Image, c.views.LayerDetails.CurrentLayer.Index, p, dst)
	}
	if err != nil {
		// an extract problem is not a reason to end the session, the full report is also shown once the TUI exits
		bus.Notify(fmt.Sprintf("extract of %s: %v", p, err))
		c.views.Status.Notify(fmt.Sprintf("extract of %s failed: %s", p, strings.ReplaceAll(err.Error(), "\n", "; ")))
	} else {
		c.views.Status.Notify(fmt.Sprintf("extracted %s to %s", p, dst))
	}
	return c.views.Status.Render()
}

func (c *controller) onFileTreeViewOptionChange() error {
	err := c.views.Status.Update()
	if err != nil {
		return err
	}
	return c.views.Status.Render()
}

func (c *controller) onFilterEdit(filter string) error {
	var filterRegex *regexp.Regexp
	var err error

	if len(filter) > 0 {
		filterRegex, err = regexp.Compile(filter)
		if err != nil {
			return err
		}
	}

	c.views.Tree.SetFilterRegex(filterRegex)

	err = c.views.Tree.Update()
	if err != nil {
		return err
	}

	return c.views.Tree.Render()
}

func (c *controller) onLayerChange(selection viewmodel.LayerSelection) error {
	// update the details
	c.views.LayerDetails.CurrentLayer = selection.Layer

	// update the filetree
	err := c.views.Tree.SetTree(selection.BottomTreeStart, selection.BottomTreeStop, selection.TopTreeStart, selection.TopTreeStop)
	if err != nil {
		return err
	}

	if c.views.Layer.CompareMode() == viewmodel.CompareAllLayers {
		c.views.Tree.SetTitle("Aggregated Layer Contents")
	} else {
		c.views.Tree.SetTitle("Current Layer Contents")
	}

	// update details and filetree panes
	return c.UpdateAndRender()
}

func (c *controller) UpdateAndRender() error {
	err := c.Update()
	if err != nil {
		return fmt.Errorf("controller failed update: %w", err)
	}

	err = c.Render()
	if err != nil {
		return fmt.Errorf("controller failed render: %w", err)
	}

	return nil
}

// Update refreshes the state objects for future rendering.
func (c *controller) Update() error {
	for _, v := range c.views.Renderers() {
		err := v.Update()
		if err != nil {
			return fmt.Errorf("controller unable to update view: %w", err)
		}
	}
	return nil
}

// Render flushes the state objects to the screen.
func (c *controller) Render() error {
	for _, v := range c.views.Renderers() {
		if v.IsVisible() {
			err := v.Render()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//nolint:dupl
func (c *controller) NextPane() (err error) {
	v := c.gui.CurrentView()
	if v == nil {
		panic("CurrentView is nil")
	}
	if v.Name() == c.views.Layer.Name() {
		_, err = c.gui.SetCurrentView(c.views.LayerDetails.Name())
		c.views.Status.SetCurrentView(c.views.LayerDetails)
	} else if v.Name() == c.views.LayerDetails.Name() {
		_, err = c.gui.SetCurrentView(c.views.ImageDetails.Name())
		c.views.Status.SetCurrentView(c.views.ImageDetails)
	} else if v.Name() == c.views.ImageDetails.Name() {
		_, err = c.gui.SetCurrentView(c.views.Layer.Name())
		c.views.Status.SetCurrentView(c.views.Layer)
	}

	if err != nil {
		return fmt.Errorf("controller unable to switch to next pane: %w", err)
	}

	return c.UpdateAndRender()
}

//nolint:dupl
func (c *controller) PrevPane() (err error) {
	v := c.gui.CurrentView()
	if v == nil {
		panic("Current view is nil")
	}
	if v.Name() == c.views.Layer.Name() {
		_, err = c.gui.SetCurrentView(c.views.ImageDetails.Name())
		c.views.Status.SetCurrentView(c.views.ImageDetails)
	} else if v.Name() == c.views.LayerDetails.Name() {
		_, err = c.gui.SetCurrentView(c.views.Layer.Name())
		c.views.Status.SetCurrentView(c.views.Layer)
	} else if v.Name() == c.views.ImageDetails.Name() {
		_, err = c.gui.SetCurrentView(c.views.LayerDetails.Name())
		c.views.Status.SetCurrentView(c.views.LayerDetails)
	}

	if err != nil {
		return fmt.Errorf("controller unable to switch to previous pane: %w", err)
	}

	return c.UpdateAndRender()
}

// ToggleView switches between the file view and the layer view and re-renders the screen.
func (c *controller) ToggleView() (err error) {
	v := c.gui.CurrentView()
	if v == nil || v.Name() == c.views.Layer.Name() {
		_, err = c.gui.SetCurrentView(c.views.Tree.Name())
		c.views.Status.SetCurrentView(c.views.Tree)
	} else {
		_, err = c.gui.SetCurrentView(c.views.Layer.Name())
		c.views.Status.SetCurrentView(c.views.Layer)
	}

	if err != nil {
		return fmt.Errorf("controller unable to toggle view: %w", err)
	}

	return c.UpdateAndRender()
}

func (c *controller) CloseFilterView() error {
	// filter view needs to be visible
	if c.views.Filter.IsVisible() {
		// toggle filter view
		return c.ToggleFilterView()
	}
	return nil
}

func (c *controller) ToggleFilterView() error {
	opening := !c.views.Filter.IsVisible()
	if opening {
		// the layout focuses the layer view before any key can be handled, so there is always a current view
		c.preFilterView, c.preFilterHelp = c.gui.CurrentView().Name(), c.views.Status.CurrentView()
	}

	// delete all user input from the tree view
	err := c.views.Filter.ToggleVisible()
	if err != nil {
		return fmt.Errorf("unable to toggle filter visibility: %w", err)
	}

	if opening {
		c.views.Status.SetCurrentView(c.views.Filter)
		return c.UpdateAndRender()
	}

	// we have just hidden the filter view, remove any filter from the tree...
	c.views.Tree.SetFilterRegex(nil)

	// ...and give focus back to the view that had it
	if _, err = c.gui.SetCurrentView(c.preFilterView); err != nil {
		return fmt.Errorf("unable to toggle filter view (back): %w", err)
	}
	c.views.Status.SetCurrentView(c.preFilterHelp)

	return c.UpdateAndRender()
}
