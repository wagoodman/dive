package controller

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/key"
	"strings"

	"github.com/jroimartin/gocui"
)

// StatusController holds the UI objects and data models for populating the bottom-most pane. Specifically the panel
// shows the user a set of possible actions to take in the window and currently selected pane.
type StatusController struct {
	name string
	gui  *gocui.Gui
	view *gocui.View

	helpKeys []*key.Binding
}

// NewStatusController creates a new view object attached the the global [gocui] screen object.
func NewStatusController(name string, gui *gocui.Gui) (controller *StatusController) {
	controller = new(StatusController)

	// populate main fields
	controller.name = name
	controller.gui = gui
	controller.helpKeys = make([]*key.Binding, 0)

	return controller
}

func (controller *StatusController) Name() string {
	return controller.name
}

func (controller *StatusController) AddHelpKeys(keys ...*key.Binding) {
	controller.helpKeys = append(controller.helpKeys, keys...)
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (controller *StatusController) Setup(v *gocui.View, header *gocui.View) error {

	// set controller options
	controller.view = v
	controller.view.Frame = false

	return controller.Render()
}

// IsVisible indicates if the status view pane is currently initialized.
func (controller *StatusController) IsVisible() bool {
	return controller != nil
}

// CursorDown moves the cursor down in the details pane (currently indicates nothing).
func (controller *StatusController) CursorDown() error {
	return nil
}

// CursorUp moves the cursor up in the details pane (currently indicates nothing).
func (controller *StatusController) CursorUp() error {
	return nil
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (controller *StatusController) Update() error {
	return nil
}

// Render flushes the state objects to the screen.
func (controller *StatusController) Render() error {
	controller.gui.Update(func(g *gocui.Gui) error {
		controller.view.Clear()
		_, err := fmt.Fprintln(controller.view, controller.KeyHelp()+format.StatusNormal("▏"+strings.Repeat(" ", 1000)))
		if err != nil {
			logrus.Debug("unable to write to buffer: ", err)
		}

		return err
	})
	return nil
}

// KeyHelp indicates all the possible global actions a user can take when any pane is selected.
func (controller *StatusController) KeyHelp() string {
	var help string
	for _, binding := range controller.helpKeys {
		help += binding.RenderKeyHelp()
	}
	return help
}
