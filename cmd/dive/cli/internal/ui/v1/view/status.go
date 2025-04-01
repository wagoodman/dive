package view

import (
	"fmt"
	"github.com/anchore/go-logger"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/format"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/key"
	"github.com/wagoodman/dive/internal/log"
	"github.com/wagoodman/dive/internal/utils"
	"strings"

	"github.com/awesome-gocui/gocui"
)

// Status holds the UI objects and data models for populating the bottom-most pane. Specifically the panel
// shows the user a set of possible actions to take in the window and currently selected pane.
type Status struct {
	name   string
	gui    *gocui.Gui
	view   *gocui.View
	logger logger.Logger

	selectedView    Helper
	requestedHeight int

	helpKeys []*key.Binding
}

// newStatusView creates a new view object attached the global [gocui] screen object.
func newStatusView(gui *gocui.Gui) *Status {
	c := new(Status)

	// populate main fields
	c.name = "status"
	c.gui = gui
	c.helpKeys = make([]*key.Binding, 0)
	c.requestedHeight = 1
	c.logger = log.Nested("ui", "status")

	return c
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
	v.logger.Trace("setup()")

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
	v.logger.Trace("render()")

	v.gui.Update(func(g *gocui.Gui) error {
		v.view.Clear()

		var selectedHelp string
		if v.selectedView != nil {
			selectedHelp = v.selectedView.KeyHelp()
		}

		_, err := fmt.Fprintln(v.view, v.KeyHelp()+selectedHelp+format.StatusNormal("▏"+strings.Repeat(" ", 1000)))
		if err != nil {
			v.logger.WithFields("error", err).Debug("unable to write to buffer")
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
	v.logger.Tracef("layout(minX: %d, minY: %d, maxX: %d, maxY: %d)", minX, minY, maxX, maxY)

	view, viewErr := g.SetView(v.Name(), minX, minY, maxX, maxY, 0)
	if utils.IsNewView(viewErr) {
		err := v.Setup(view)
		if err != nil {
			return fmt.Errorf("unable to setup status controller: %w", err)
		}
	}
	return nil
}

func (v *Status) RequestedSize(available int) *int {
	return &v.requestedHeight
}
