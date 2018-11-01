package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"strings"
)

// DetailsView holds the UI objects and data models for populating the bottom-most pane. Specifcially the panel
// shows the user a set of possible actions to take in the window and currently selected pane.
type StatusView struct {
	Name string
	gui  *gocui.Gui
	view *gocui.View
}

// NewStatusView creates a new view object attached the the global [gocui] screen object.
func NewStatusView(name string, gui *gocui.Gui) (statusView *StatusView) {
	statusView = new(StatusView)

	// populate main fields
	statusView.Name = name
	statusView.gui = gui

	return statusView
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (view *StatusView) Setup(v *gocui.View, header *gocui.View) error {

	// set view options
	view.view = v
	view.view.Frame = false

	view.Render()

	return nil
}

// IsVisible indicates if the status view pane is currently initialized.
func (view *StatusView) IsVisible() bool {
	if view == nil {
		return false
	}
	return true
}

// CursorDown moves the cursor down in the details pane (currently indicates nothing).
func (view *StatusView) CursorDown() error {
	return nil
}

// CursorUp moves the cursor up in the details pane (currently indicates nothing).
func (view *StatusView) CursorUp() error {
	return nil
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (view *StatusView) Update() error {
	return nil
}

// Render flushes the state objects to the screen.
func (view *StatusView) Render() error {
	view.gui.Update(func(g *gocui.Gui) error {
		view.view.Clear()
		fmt.Fprintln(view.view, view.KeyHelp()+Views.lookup[view.gui.CurrentView().Name()].KeyHelp()+Formatting.StatusNormal("▏"+strings.Repeat(" ", 1000)))

		return nil
	})
	// todo: blerg
	return nil
}

// KeyHelp indicates all the possible global actions a user can take when any pane is selected.
func (view *StatusView) KeyHelp() string {
	return renderStatusOption(GlobalKeybindings.quit[0].String(), "Quit", false) +
		renderStatusOption(GlobalKeybindings.toggleView[0].String(), "Switch view", false) +
		renderStatusOption(GlobalKeybindings.filterView[0].String(), "Filter files", Views.Filter.IsVisible())
}
