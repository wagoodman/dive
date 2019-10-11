package ui

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/key"
	"strings"

	"github.com/jroimartin/gocui"
)

// statusController holds the UI objects and data models for populating the bottom-most pane. Specifically the panel
// shows the user a set of possible actions to take in the window and currently selected pane.
type statusController struct {
	name string
	gui  *gocui.Gui
	view *gocui.View

	helpKeys []*key.Binding
}

// newStatusController creates a new view object attached the the global [gocui] screen object.
func newStatusController(name string, gui *gocui.Gui) (controller *statusController) {
	controller = new(statusController)

	// populate main fields
	controller.name = name
	controller.gui = gui
	controller.helpKeys = make([]*key.Binding, 0)

	return controller
}

func (controller *statusController) AddHelpKeys(keys ...*key.Binding) {
	for _, k := range keys {
		controller.helpKeys = append(controller.helpKeys, k)
	}
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (controller *statusController) Setup(v *gocui.View, header *gocui.View) error {

	// set controller options
	controller.view = v
	controller.view.Frame = false

	return controller.Render()
}

// IsVisible indicates if the status view pane is currently initialized.
func (controller *statusController) IsVisible() bool {
	return controller != nil
}

// CursorDown moves the cursor down in the details pane (currently indicates nothing).
func (controller *statusController) CursorDown() error {
	return nil
}

// CursorUp moves the cursor up in the details pane (currently indicates nothing).
func (controller *statusController) CursorUp() error {
	return nil
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (controller *statusController) Update() error {
	return nil
}

// Render flushes the state objects to the screen.
func (controller *statusController) Render() error {
	controller.gui.Update(func(g *gocui.Gui) error {
		controller.view.Clear()
		_, err := fmt.Fprintln(controller.view, controller.KeyHelp()+controllers.lookup[controller.gui.CurrentView().Name()].KeyHelp()+format.StatusNormal("▏"+strings.Repeat(" ", 1000)))
		if err != nil {
			logrus.Debug("unable to write to buffer: ", err)
		}

		return err
	})
	return nil
}

// KeyHelp indicates all the possible global actions a user can take when any pane is selected.
func (controller *statusController) KeyHelp() string {
	var help string
	for _, binding := range controller.helpKeys {
		help += binding.RenderKeyHelp()
	}
	return help
}
