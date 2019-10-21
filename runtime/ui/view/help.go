package view

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/key"
	"strings"

	"github.com/jroimartin/gocui"
)

// Help holds the UI objects and data models for populating the bottom-most pane. Specifically the panel
// shows the user a set of possible actions to take in the window and currently selected pane.
type Help struct {
	name string
	gui  *gocui.Gui
	view *gocui.View

	selectedView View

	helpKeys []*key.Binding
}

// NewHelpView creates a new view object attached the the global [gocui] screen object.
func NewHelpView(name string, gui *gocui.Gui) (controller *Help) {
	controller = new(Help)

	// populate main fields
	controller.name = name
	controller.gui = gui
	controller.helpKeys = make([]*key.Binding, 0)

	return controller
}

func (c *Help) SetCurrentView(r View) {
	c.selectedView = r
}

func (c *Help) Height() int {
	return 1
}

func (c *Help) Width() int {
	return WidthFull
}

func (c *Help) Name() string {
	return c.name
}

func (c *Help) AddHelpKeys(keys ...*key.Binding) {
	c.helpKeys = append(c.helpKeys, keys...)
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (c *Help) Setup(v *gocui.View, header *gocui.View) error {

	// set controller options
	c.view = v
	c.view.Frame = false

	return c.Render()
}

// IsVisible indicates if the status view pane is currently initialized.
func (c *Help) IsVisible() bool {
	return c != nil
}

// CursorDown moves the cursor down in the details pane (currently indicates nothing).
func (c *Help) CursorDown() error {
	return nil
}

// CursorUp moves the cursor up in the details pane (currently indicates nothing).
func (c *Help) CursorUp() error {
	return nil
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (c *Help) Update() error {
	return nil
}

// Render flushes the state objects to the screen.
func (c *Help) Render() error {
	c.gui.Update(func(g *gocui.Gui) error {
		c.view.Clear()

		var selectedHelp string
		if c.selectedView != nil {
			selectedHelp = c.selectedView.KeyHelp()
		}

		_, err := fmt.Fprintln(c.view, c.KeyHelp()+selectedHelp+format.StatusNormal("▏"+strings.Repeat(" ", 1000)))
		if err != nil {
			logrus.Debug("unable to write to buffer: ", err)
		}

		return err
	})
	return nil
}

// KeyHelp indicates all the possible global actions a user can take when any pane is selected.
func (c *Help) KeyHelp() string {
	var help string
	for _, binding := range c.helpKeys {
		help += binding.RenderKeyHelp()
	}
	return help
}
