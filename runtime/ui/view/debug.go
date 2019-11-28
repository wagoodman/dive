package view

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/utils"
)

// Debug is just for me :)
type Debug struct {
	name   string
	gui    *gocui.Gui
	view   *gocui.View
	header *gocui.View

	selectedView Helper
}

// newDebugView creates a new view object attached the the global [gocui] screen object.
func newDebugView(gui *gocui.Gui) (controller *Debug) {
	controller = new(Debug)

	// populate main fields
	controller.name = "debug"
	controller.gui = gui

	return controller
}

func (v *Debug) SetCurrentView(r Helper) {
	v.selectedView = r
}

func (v *Debug) Name() string {
	return v.name
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (v *Debug) Setup(view *gocui.View, header *gocui.View) error {
	logrus.Tracef("view.Setup() %s", v.Name())

	// set controller options
	v.view = view
	v.view.Editable = false
	v.view.Wrap = false
	v.view.Frame = false

	v.header = header
	v.header.Editable = false
	v.header.Wrap = false
	v.header.Frame = false

	return v.Render()
}

// IsVisible indicates if the status view pane is currently initialized.
func (v *Debug) IsVisible() bool {
	return v != nil
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (v *Debug) Update() error {
	return nil
}

// OnLayoutChange is called whenever the screen dimensions are changed
func (v *Debug) OnLayoutChange() error {
	err := v.Update()
	if err != nil {
		return err
	}
	return v.Render()
}

// Render flushes the state objects to the screen.
func (v *Debug) Render() error {
	logrus.Tracef("view.Render() %s", v.Name())

	v.gui.Update(func(g *gocui.Gui) error {
		// update header...
		v.header.Clear()
		width, _ := g.Size()
		headerStr := format.RenderHeader("Debug", width, false)
		_, _ = fmt.Fprintln(v.header, headerStr)

		// update view...
		v.view.Clear()
		_, err := fmt.Fprintln(v.view, "blerg")
		if err != nil {
			logrus.Debug("unable to write to buffer: ", err)
		}

		return nil
	})
	return nil
}

func (v *Debug) Layout(g *gocui.Gui, minX, minY, maxX, maxY int) error {
	logrus.Tracef("view.Layout(minX: %d, minY: %d, maxX: %d, maxY: %d) %s", minX, minY, maxX, maxY, v.Name())

	// header
	headerSize := 1
	// note: maxY needs to account for the (invisible) border, thus a +1
	header, headerErr := g.SetView(v.Name()+"header", minX, minY, maxX, minY+headerSize+1)
	// we are going to overlap the view over the (invisible) border (so minY will be one less than expected).
	// additionally, maxY will be bumped by one to include the border
	view, viewErr := g.SetView(v.Name(), minX, minY+headerSize, maxX, maxY+1)
	if utils.IsNewView(viewErr, headerErr) {
		err := v.Setup(view, header)
		if err != nil {
			logrus.Error("unable to setup debug controller", err)
			return err
		}
	}
	return nil
}

func (v *Debug) RequestedSize(available int) *int {
	return nil
}
