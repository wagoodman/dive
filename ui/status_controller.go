package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"strings"
)

// StatusController holds the UI objects and data models for populating the bottom-most pane. Specifically the panel
// shows the user a set of possible actions to take in the window and currently selected pane.
type StatusController struct {
	Name string
	gui  *gocui.Gui
	view *gocui.View
}

// NewStatusController creates a new view object attached the the global [gocui] screen object.
func NewStatusController(name string, gui *gocui.Gui) (controller *StatusController) {
	controller = new(StatusController)

	// populate main fields
	controller.Name = name
	controller.gui = gui

	return controller
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (controller *StatusController) Setup(v *gocui.View, header *gocui.View) error {

	// set controller options
	controller.view = v
	controller.view.Frame = false

	controller.Render()

	return nil
}

// IsVisible indicates if the status view pane is currently initialized.
func (controller *StatusController) IsVisible() bool {
	if controller == nil {
		return false
	}
	return true
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
		fmt.Fprintln(controller.view, controller.KeyHelp()+Controllers.lookup[controller.gui.CurrentView().Name()].KeyHelp()+Formatting.StatusNormal("▏"+strings.Repeat(" ", 1000)))

		return nil
	})
	// todo: blerg
	return nil
}

// KeyHelp indicates all the possible global actions a user can take when any pane is selected.
func (controller *StatusController) KeyHelp() string {
	return renderStatusOption(GlobalKeybindings.quit[0].String(), "Quit", false) +
		renderStatusOption(GlobalKeybindings.toggleView[0].String(), "Switch view", false) +
		renderStatusOption(GlobalKeybindings.filterView[0].String(), "Filter", Controllers.Filter.IsVisible())
}
