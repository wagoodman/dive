package view

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/format"
	"strings"
)

type FilterEditListener func(string) error

// Filter holds the UI objects and data models for populating the bottom row. Specifically the pane that
// allows the user to filter the file tree by path.
type Filter struct {
	name      string
	gui       *gocui.Gui
	view      *gocui.View
	header    *gocui.View
	headerStr string
	maxLength int
	hidden    bool

	filterEditListeners []FilterEditListener
}

// NewFilterView creates a new view object attached the the global [gocui] screen object.
func NewFilterView(name string, gui *gocui.Gui) (controller *Filter) {
	controller = new(Filter)

	controller.filterEditListeners = make([]FilterEditListener, 0)

	// populate main fields
	controller.name = name
	controller.gui = gui
	controller.headerStr = "Path Filter: "
	controller.hidden = true

	return controller
}

func (c *Filter) AddFilterEditListener(listener ...FilterEditListener) {
	c.filterEditListeners = append(c.filterEditListeners, listener...)
}

func (c *Filter) Name() string {
	return c.name
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (c *Filter) Setup(v *gocui.View, header *gocui.View) error {

	// set controller options
	c.view = v
	c.maxLength = 200
	c.view.Frame = false
	c.view.BgColor = gocui.AttrReverse
	c.view.Editable = true
	c.view.Editor = c

	c.header = header
	c.header.BgColor = gocui.AttrReverse
	c.header.Editable = false
	c.header.Wrap = false
	c.header.Frame = false

	return c.Render()
}

// ToggleFilterView shows/hides the file tree filter pane.
func (c *Filter) ToggleVisible() error {
	// delete all user input from the tree view
	c.view.Clear()

	// toggle hiding
	c.hidden = !c.hidden

	if !c.hidden {
		_, err := c.gui.SetCurrentView(c.name)
		if err != nil {
			logrus.Error("unable to toggle filter view: ", err)
			return err
		}
		return nil
	}

	// reset the cursor for the next time it is visible
	// Note: there is a subtle gocui behavior here where this cannot be called when the view
	// is newly visible. Is this a problem with dive or gocui?
	return c.view.SetCursor(0, 0)
}

// todo: remove the need for this
func (c *Filter) HeaderStr() string {
	return c.headerStr
}

// IsVisible indicates if the filter view pane is currently initialized
func (c *Filter) IsVisible() bool {
	if c == nil {
		return false
	}
	return !c.hidden
}

// CursorDown moves the cursor down in the filter pane (currently indicates nothing).
func (c *Filter) CursorDown() error {
	return nil
}

// CursorUp moves the cursor up in the filter pane (currently indicates nothing).
func (c *Filter) CursorUp() error {
	return nil
}

// Edit intercepts the key press events in the filer view to update the file view in real time.
func (c *Filter) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if !c.IsVisible() {
		return
	}

	cx, _ := v.Cursor()
	ox, _ := v.Origin()
	limit := ox+cx+1 > c.maxLength
	switch {
	case ch != 0 && mod == 0 && !limit:
		v.EditWrite(ch)
	case key == gocui.KeySpace && !limit:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	}

	// notify listeners
	c.notifyFilterEditListeners()
}

func (c *Filter) notifyFilterEditListeners() {
	currentValue := strings.TrimSpace(c.view.Buffer())
	for _, listener := range c.filterEditListeners {
		err := listener(currentValue)
		if err != nil {
			// note: cannot propagate error from here since this is from the main gogui thread
			logrus.Errorf("notifyFilterEditListeners: %+v", err)
		}
	}
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (c *Filter) Update() error {
	return nil
}

// Render flushes the state objects to the screen. Currently this is the users path filter input.
func (c *Filter) Render() error {
	c.gui.Update(func(g *gocui.Gui) error {
		_, err := fmt.Fprintln(c.header, format.Header(c.headerStr))
		if err != nil {
			logrus.Error("unable to write to buffer: ", err)
		}
		return err
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected.
func (c *Filter) KeyHelp() string {
	return format.StatusControlNormal("▏Type to filter the file tree ")
}
