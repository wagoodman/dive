package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/format"
)

// filterController holds the UI objects and data models for populating the bottom row. Specifically the pane that
// allows the user to filter the file tree by path.
type filterController struct {
	name      string
	gui       *gocui.Gui
	view      *gocui.View
	header    *gocui.View
	headerStr string
	maxLength int
	hidden    bool
}

// newFilterController creates a new view object attached the the global [gocui] screen object.
func newFilterController(name string, gui *gocui.Gui) (controller *filterController) {
	controller = new(filterController)

	// populate main fields
	controller.name = name
	controller.gui = gui
	controller.headerStr = "Path Filter: "
	controller.hidden = true

	return controller
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (controller *filterController) Setup(v *gocui.View, header *gocui.View) error {

	// set controller options
	controller.view = v
	controller.maxLength = 200
	controller.view.Frame = false
	controller.view.BgColor = gocui.AttrReverse
	controller.view.Editable = true
	controller.view.Editor = controller

	controller.header = header
	controller.header.BgColor = gocui.AttrReverse
	controller.header.Editable = false
	controller.header.Wrap = false
	controller.header.Frame = false

	return controller.Render()
}

// IsVisible indicates if the filter view pane is currently initialized
func (controller *filterController) IsVisible() bool {
	if controller == nil {
		return false
	}
	return !controller.hidden
}

// CursorDown moves the cursor down in the filter pane (currently indicates nothing).
func (controller *filterController) CursorDown() error {
	return nil
}

// CursorUp moves the cursor up in the filter pane (currently indicates nothing).
func (controller *filterController) CursorUp() error {
	return nil
}

// Edit intercepts the key press events in the filer view to update the file view in real time.
func (controller *filterController) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if !controller.IsVisible() {
		return
	}

	cx, _ := v.Cursor()
	ox, _ := v.Origin()
	limit := ox+cx+1 > controller.maxLength
	switch {
	case ch != 0 && mod == 0 && !limit:
		v.EditWrite(ch)
	case key == gocui.KeySpace && !limit:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	}
	if controllers.Tree != nil {
		_ = controllers.Tree.Update()
		_ = controllers.Tree.Render()
	}
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (controller *filterController) Update() error {
	return nil
}

// Render flushes the state objects to the screen. Currently this is the users path filter input.
func (controller *filterController) Render() error {
	controller.gui.Update(func(g *gocui.Gui) error {
		_, err := fmt.Fprintln(controller.header, format.Header(controller.headerStr))
		if err != nil {
			logrus.Error("unable to write to buffer: ", err)
		}
		return err
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected.
func (controller *filterController) KeyHelp() string {
	return format.StatusControlNormal("▏Type to filter the file tree ")
}
