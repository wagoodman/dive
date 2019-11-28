package view

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/key"
	"github.com/wagoodman/dive/utils"
	"strings"

	"github.com/jroimartin/gocui"
)

// Status holds the UI objects and data models for populating the bottom-most pane. Specifically the panel
// shows the user a set of possible actions to take in the window and currently selected pane.
type Status struct {
	name string
	gui  *gocui.Gui
	view *gocui.View

	selectedView    Helper
	requestedHeight int

	helpKeys []*key.Binding
}

// newStatusView creates a new view object attached the the global [gocui] screen object.
func newStatusView(gui *gocui.Gui) (controller *Status) {
	controller = new(Status)

	// populate main fields
	controller.name = "status"
	controller.gui = gui
	controller.helpKeys = make([]*key.Binding, 0)
	controller.requestedHeight = 1

	return controller
}

func (v *Status) SetCurrentView(r Helper) {
	v.selectedView = r
}

func (v *Status) Name() string {
	return v.name
}

func (v *Status) AddHelpKeys(keys ...*key.Binding) {
	v.helpKeys = append(v.helpKeys, keys...)
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (v *Status) Setup(view *gocui.View) error {
	logrus.Tracef("view.Setup() %s", v.Name())

	// set controller options
	v.view = view
	v.view.Frame = false

	return v.Render()
}

// IsVisible indicates if the status view pane is currently initialized.
func (v *Status) IsVisible() bool {
	return v != nil
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (v *Status) Update() error {
	return nil
}

// OnLayoutChange is called whenever the screen dimensions are changed
func (v *Status) OnLayoutChange() error {
	err := v.Update()
	if err != nil {
		return err
	}
	return v.Render()
}

// Render flushes the state objects to the screen.
func (v *Status) Render() error {
	logrus.Tracef("view.Render() %s", v.Name())

	v.gui.Update(func(g *gocui.Gui) error {
		v.view.Clear()

		var selectedHelp string
		if v.selectedView != nil {
			selectedHelp = v.selectedView.KeyHelp()
		}

		_, err := fmt.Fprintln(v.view, v.KeyHelp()+selectedHelp+format.StatusNormal("▏"+strings.Repeat(" ", 1000)))
		if err != nil {
			logrus.Debug("unable to write to buffer: ", err)
		}

		return err
	})
	return nil
}

// KeyHelp indicates all the possible global actions a user can take when any pane is selected.
func (v *Status) KeyHelp() string {
	var help string
	for _, binding := range v.helpKeys {
		help += binding.RenderKeyHelp()
	}
	return help
}

func (v *Status) Layout(g *gocui.Gui, minX, minY, maxX, maxY int) error {
	logrus.Tracef("view.Layout(minX: %d, minY: %d, maxX: %d, maxY: %d) %s", minX, minY, maxX, maxY, v.Name())

	view, viewErr := g.SetView(v.Name(), minX, minY, maxX, maxY)
	if utils.IsNewView(viewErr) {
		err := v.Setup(view)
		if err != nil {
			logrus.Error("unable to setup status controller", err)
			return err
		}
	}
	return nil
}

func (v *Status) RequestedSize(available int) *int {
	return &v.requestedHeight
}
