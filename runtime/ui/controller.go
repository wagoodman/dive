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
	gui     *gocui.Gui
	Tree    *view.FileTree
	Layer   *view.Layer
	Status  *view.Status
	Filter  *view.Filter
	Details *view.Details
	lookup  map[string]view.Renderer
}

func NewCollection(g *gocui.Gui, analysis *image.AnalysisResult, cache filetree.Comparer) (*Controller, error) {
	var err error

	controller := &Controller{
		gui: g,
	}
	controller.lookup = make(map[string]view.Renderer)

	controller.Layer, err = view.NewLayerView("layers", g, analysis.Layers)
	if err != nil {
		return nil, err
	}
	controller.lookup[controller.Layer.Name()] = controller.Layer

	//treeStack, err := filetree.StackTreeRange(analysis.RefTrees, 0, 0)
	//if err != nil {
	//	return nil, err
	//}
	treeStack := analysis.RefTrees[0]
	controller.Tree, err = view.NewFileTreeView("filetree", g, treeStack, analysis.RefTrees, cache)
	if err != nil {
		return nil, err
	}
	controller.lookup[controller.Tree.Name()] = controller.Tree

	// layer view cursor down event should trigger an update in the file tree
	controller.Layer.AddLayerChangeListener(controller.onLayerChange)

	controller.Status = view.NewStatusView("status", g)
	controller.lookup[controller.Status.Name()] = controller.Status
	// set the layer view as the first selected view
	controller.Status.SetCurrentView(controller.Layer)

	// update the status pane when a filetree option is changed by the user
	controller.Tree.AddViewOptionChangeListener(controller.onFileTreeViewOptionChange)

	controller.Filter = view.NewFilterView("filter", g)
	controller.lookup[controller.Filter.Name()] = controller.Filter
	controller.Filter.AddFilterEditListener(controller.onFilterEdit)

	controller.Details = view.NewDetailsView("details", g, analysis.Efficiency, analysis.Inefficiencies, analysis.SizeBytes)
	controller.lookup[controller.Details.Name()] = controller.Details

	// propagate initial conditions to necessary views
	err = controller.onLayerChange(viewmodel.LayerSelection{
		Layer:           controller.Layer.CurrentLayer(),
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
	err := c.Status.Update()
	if err != nil {
		return err
	}
	return c.Status.Render()
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

	c.Tree.SetFilterRegex(filterRegex)

	err = c.Tree.Update()
	if err != nil {
		return err
	}

	return c.Tree.Render()
}

func (c *Controller) onLayerChange(selection viewmodel.LayerSelection) error {
	// update the details
	c.Details.SetCurrentLayer(selection.Layer)

	// update the filetree
	err := c.Tree.SetTree(selection.BottomTreeStart, selection.BottomTreeStop, selection.TopTreeStart, selection.TopTreeStop)
	if err != nil {
		return err
	}

	if c.Layer.CompareMode == view.CompareAll {
		c.Tree.SetTitle("Aggregated Layer Contents")
	} else {
		c.Tree.SetTitle("Current Layer Contents")
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
	for _, controller := range c.lookup {
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
	for _, controller := range c.lookup {
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
	if v == nil || v.Name() == c.Layer.Name() {
		_, err = c.gui.SetCurrentView(c.Tree.Name())
		c.Status.SetCurrentView(c.Tree)
	} else {
		_, err = c.gui.SetCurrentView(c.Layer.Name())
		c.Status.SetCurrentView(c.Layer)
	}

	if err != nil {
		logrus.Error("unable to toggle view: ", err)
		return err
	}

	return c.UpdateAndRender()
}

func (c *Controller) ToggleFilterView() error {
	// delete all user input from the tree view
	err := c.Filter.ToggleVisible()
	if err != nil {
		logrus.Error("unable to toggle filter visibility: ", err)
		return err
	}

	// we have just hidden the filter view...
	if !c.Filter.IsVisible() {
		// ...remove any filter from the tree
		c.Tree.SetFilterRegex(nil)

		// ...adjust focus to a valid (visible) view
		err = c.ToggleView()
		if err != nil {
			logrus.Error("unable to toggle filter view (back): ", err)
			return err
		}
	}

	return c.UpdateAndRender()
}
