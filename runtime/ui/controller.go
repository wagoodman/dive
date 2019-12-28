package ui

import (
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/view"
	"github.com/wagoodman/dive/runtime/ui/viewmodel"
	"regexp"
)

type Controller struct {
	gui   *gocui.Gui
	views *view.Views
}

func NewCollection(g *gocui.Gui, analysis *image.AnalysisResult, cache filetree.Comparer) (*Controller, error) {
	views, err := view.NewViews(g, analysis, cache)
	if err != nil {
		return nil, err
	}

	controller := &Controller{
		gui:   g,
		views: views,
	}

	// layer view cursor down event should trigger an update in the file tree
	controller.views.Layer.AddLayerChangeListener(controller.onLayerChange)

	// update the status pane when a filetree option is changed by the user
	controller.views.Tree.AddViewOptionChangeListener(controller.onFileTreeViewOptionChange)

	// update the tree view while the user types into the filter view
	controller.views.Filter.AddFilterEditListener(controller.onFilterEdit)

	// propagate initial conditions to necessary views
	err = controller.onLayerChange(viewmodel.LayerSelection{
		Layer:           controller.views.Layer.CurrentLayer(),
		BottomTreeStart: 0,
		BottomTreeStop:  0,
		TopTreeStart:    0,
		TopTreeStop:     0,
	})

	if err != nil {
		return nil, err
	}

	return controller, nil
}

func (c *Controller) onFileTreeViewOptionChange() error {
	err := c.views.Status.Update()
	if err != nil {
		return err
	}
	return c.views.Status.Render()
}

func (c *Controller) onFilterEdit(filter string) error {
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

func (c *Controller) onLayerChange(selection viewmodel.LayerSelection) error {
	// update the details
	c.views.Details.SetCurrentLayer(selection.Layer)

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

func (c *Controller) UpdateAndRender() error {
	err := c.Update()
	if err != nil {
		logrus.Debug("failed update: ", err)
		return err
	}

	err = c.Render()
	if err != nil {
		logrus.Debug("failed render: ", err)
		return err
	}

	return nil
}

// Update refreshes the state objects for future rendering.
func (c *Controller) Update() error {
	for _, controller := range c.views.All() {
		err := controller.Update()
		if err != nil {
			logrus.Debug("unable to update controller: ")
			return err
		}
	}
	return nil
}

// Render flushes the state objects to the screen.
func (c *Controller) Render() error {
	for _, controller := range c.views.All() {
		if controller.IsVisible() {
			err := controller.Render()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ToggleView switches between the file view and the layer view and re-renders the screen.
func (c *Controller) ToggleView() (err error) {
	v := c.gui.CurrentView()
	if v == nil || v.Name() == c.views.Layer.Name() {
		_, err = c.gui.SetCurrentView(c.views.Tree.Name())
		c.views.Status.SetCurrentView(c.views.Tree)
	} else {
		_, err = c.gui.SetCurrentView(c.views.Layer.Name())
		c.views.Status.SetCurrentView(c.views.Layer)
	}

	if err != nil {
		logrus.Error("unable to toggle view: ", err)
		return err
	}

	return c.UpdateAndRender()
}

func (c *Controller) ToggleFilterView() error {
	// delete all user input from the tree view
	err := c.views.Filter.ToggleVisible()
	if err != nil {
		logrus.Error("unable to toggle filter visibility: ", err)
		return err
	}

	// we have just hidden the filter view...
	if !c.views.Filter.IsVisible() {
		// ...remove any filter from the tree
		c.views.Tree.SetFilterRegex(nil)

		// ...adjust focus to a valid (visible) view
		err = c.ToggleView()
		if err != nil {
			logrus.Error("unable to toggle filter view (back): ", err)
			return err
		}
	}

	return c.UpdateAndRender()
}
