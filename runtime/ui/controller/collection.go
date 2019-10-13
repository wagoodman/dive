package controller

import (
	"errors"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
)

// var ccOnce sync.Once
var controllers *Collection

type Collection struct {
	gui     *gocui.Gui
	Tree    *FileTree
	Layer   *Layer
	Status  *Status
	Filter  *Filter
	Details *Details
	lookup  map[string]Controller
}

func NewCollection(g *gocui.Gui, analysis *image.AnalysisResult, cache filetree.TreeCache) (*Collection, error) {
	var err error

	controllers = &Collection{
		gui: g,
	}
	controllers.lookup = make(map[string]Controller)

	controllers.Layer, err = NewLayerController("layers", g, analysis.Layers)
	if err != nil {
		return nil, err
	}
	controllers.lookup[controllers.Layer.name] = controllers.Layer

	treeStack, err := filetree.StackTreeRange(analysis.RefTrees, 0, 0)
	if err != nil {
		return nil, err
	}
	controllers.Tree, err = NewFileTreeController("filetree", g, treeStack, analysis.RefTrees, cache)
	if err != nil {
		return nil, err
	}
	controllers.lookup[controllers.Tree.name] = controllers.Tree

	controllers.Status = NewStatusController("status", g)
	controllers.lookup[controllers.Status.name] = controllers.Status

	controllers.Filter = NewFilterController("filter", g)
	controllers.lookup[controllers.Filter.name] = controllers.Filter

	controllers.Details = NewDetailsController("details", g, analysis.Efficiency, analysis.Inefficiencies)
	controllers.lookup[controllers.Details.name] = controllers.Details
	return controllers, nil
}

func (c *Collection) UpdateAndRender() error {
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
func (c *Collection) Update() error {
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
func (c *Collection) Render() error {
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
func (c *Collection) ToggleView() (err error) {
	v := c.gui.CurrentView()
	if v == nil || v.Name() == c.Layer.Name() {
		_, err = c.gui.SetCurrentView(c.Tree.Name())
	} else {
		_, err = c.gui.SetCurrentView(c.Layer.Name())
	}

	if err != nil {
		logrus.Error("unable to toggle view: ", err)
		return err
	}

	return c.UpdateAndRender()
}

func (c *Collection) ToggleFilterView() error {
	// delete all user input from the tree view
	err := c.Filter.ToggleVisible()
	if err != nil {
		logrus.Error("unable to toggle filter visibility: ", err)
		return err
	}

	// we have just hidden the filter view, adjust focus to a valid (visible) view
	if !c.Filter.IsVisible() {
		err = c.ToggleView()
		if err != nil {
			logrus.Error("unable to toggle filter view (back): ", err)
			return err
		}
	}

	return c.UpdateAndRender()
}

// CursorDown moves the cursor down in the currently selected gocui pane, scrolling the screen as needed.
func (c *Collection) CursorDown(g *gocui.Gui, v *gocui.View) error {
	return c.CursorStep(g, v, 1)
}

// CursorUp moves the cursor up in the currently selected gocui pane, scrolling the screen as needed.
func (c *Collection) CursorUp(g *gocui.Gui, v *gocui.View) error {
	return c.CursorStep(g, v, -1)
}

// Moves the cursor the given step distance, setting the origin to the new cursor line
func (c *Collection) CursorStep(g *gocui.Gui, v *gocui.View, step int) error {
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
