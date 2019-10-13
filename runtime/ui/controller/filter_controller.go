package controller

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/format"
)

// FilterController holds the UI objects and data models for populating the bottom row. Specifically the pane that
// allows the user to filter the file tree by path.
type FilterController struct {
	name      string
	gui       *gocui.Gui
	view      *gocui.View
	header    *gocui.View
	headerStr string
	maxLength int
	hidden    bool
}

// NewFilterController creates a new view object attached the the global [gocui] screen object.
func NewFilterController(name string, gui *gocui.Gui) (controller *FilterController) {
	controller = new(FilterController)

	// populate main fields
	controller.name = name
	controller.gui = gui
	controller.headerStr = "Path Filter: "
	controller.hidden = true

	return controller
}

func (controller *FilterController) Name() string {
	return controller.name
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (controller *FilterController) Setup(v *gocui.View, header *gocui.View) error {

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

// ToggleFilterView shows/hides the file tree filter pane.
func (controller *FilterController) ToggleVisible() error {
	// delete all user input from the tree view
	controller.view.Clear()

	// toggle hiding
	controller.hidden = !controller.hidden

	if !controller.hidden {
		_, err := controller.gui.SetCurrentView(controller.name)
		if err != nil {
			logrus.Error("unable to toggle filter view: ", err)
			return err
		}
		return controllers.UpdateAndRender()
	}

	// reset the cursor for the next time it is visible
	// Note: there is a subtle gocui behavior here where this cannot be called when the view
	// is newly visible. Is this a problem with dive or gocui?
	return controller.view.SetCursor(0, 0)
}

// todo: remove the need for this
func (controller *FilterController) HeaderStr() string {
	return controller.headerStr
}

// IsVisible indicates if the filter view pane is currently initialized
func (controller *FilterController) IsVisible() bool {
	if controller == nil {
		return false
	}
	return !controller.hidden
}

// CursorDown moves the cursor down in the filter pane (currently indicates nothing).
func (controller *FilterController) CursorDown() error {
	return nil
}

// CursorUp moves the cursor up in the filter pane (currently indicates nothing).
func (controller *FilterController) CursorUp() error {
	return nil
}

// Edit intercepts the key press events in the filer view to update the file view in real time.
func (controller *FilterController) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
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
func (controller *FilterController) Update() error {
	return nil
}

// Render flushes the state objects to the screen. Currently this is the users path filter input.
func (controller *FilterController) Render() error {
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
func (controller *FilterController) KeyHelp() string {
	return format.StatusControlNormal("▏Type to filter the file tree ")
}
