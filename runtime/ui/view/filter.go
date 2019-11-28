package view

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/utils"
	"strings"
)

type FilterEditListener func(string) error

// Filter holds the UI objects and data models for populating the bottom row. Specifically the pane that
// allows the user to filter the file tree by path.
type Filter struct {
	name            string
	gui             *gocui.Gui
	view            *gocui.View
	header          *gocui.View
	labelStr        string
	maxLength       int
	hidden          bool
	requestedHeight int

	filterEditListeners []FilterEditListener
}

// newFilterView creates a new view object attached the the global [gocui] screen object.
func newFilterView(gui *gocui.Gui) (controller *Filter) {
	controller = new(Filter)

	controller.filterEditListeners = make([]FilterEditListener, 0)

	// populate main fields
	controller.name = "filter"
	controller.gui = gui
	controller.labelStr = "Path Filter: "
	controller.hidden = true

	controller.requestedHeight = 1

	return controller
}

func (v *Filter) AddFilterEditListener(listener ...FilterEditListener) {
	v.filterEditListeners = append(v.filterEditListeners, listener...)
}

func (v *Filter) Name() string {
	return v.name
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (v *Filter) Setup(view *gocui.View, header *gocui.View) error {
	logrus.Tracef("view.Setup() %s", v.Name())

	// set controller options
	v.view = view
	v.maxLength = 200
	v.view.Frame = false
	v.view.BgColor = gocui.AttrReverse
	v.view.Editable = true
	v.view.Editor = v

	v.header = header
	v.header.BgColor = gocui.AttrReverse
	v.header.Editable = false
	v.header.Wrap = false
	v.header.Frame = false

	return v.Render()
}

// ToggleFilterView shows/hides the file tree filter pane.
func (v *Filter) ToggleVisible() error {
	// delete all user input from the tree view
	v.view.Clear()

	// toggle hiding
	v.hidden = !v.hidden

	if !v.hidden {
		_, err := v.gui.SetCurrentView(v.name)
		if err != nil {
			logrus.Error("unable to toggle filter view: ", err)
			return err
		}
		return nil
	}

	// reset the cursor for the next time it is visible
	// Note: there is a subtle gocui behavior here where this cannot be called when the view
	// is newly visible. Is this a problem with dive or gocui?
	return v.view.SetCursor(0, 0)
}

// IsVisible indicates if the filter view pane is currently initialized
func (v *Filter) IsVisible() bool {
	if v == nil {
		return false
	}
	return !v.hidden
}

// Edit intercepts the key press events in the filer view to update the file view in real time.
func (v *Filter) Edit(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if !v.IsVisible() {
		return
	}

	cx, _ := view.Cursor()
	ox, _ := view.Origin()
	limit := ox+cx+1 > v.maxLength
	switch {
	case ch != 0 && mod == 0 && !limit:
		view.EditWrite(ch)
	case key == gocui.KeySpace && !limit:
		view.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		view.EditDelete(true)
	}

	// notify listeners
	v.notifyFilterEditListeners()
}

func (v *Filter) notifyFilterEditListeners() {
	currentValue := strings.TrimSpace(v.view.Buffer())
	for _, listener := range v.filterEditListeners {
		err := listener(currentValue)
		if err != nil {
			// note: cannot propagate error from here since this is from the main gogui thread
			logrus.Errorf("notifyFilterEditListeners: %+v", err)
		}
	}
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (v *Filter) Update() error {
	return nil
}

// Render flushes the state objects to the screen. Currently this is the users path filter input.
func (v *Filter) Render() error {
	logrus.Tracef("view.Render() %s", v.Name())

	v.gui.Update(func(g *gocui.Gui) error {
		_, err := fmt.Fprintln(v.header, format.Header(v.labelStr))
		if err != nil {
			logrus.Error("unable to write to buffer: ", err)
		}
		return err
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected.
func (v *Filter) KeyHelp() string {
	return format.StatusControlNormal("▏Type to filter the file tree ")
}

// OnLayoutChange is called whenever the screen dimensions are changed
func (v *Filter) OnLayoutChange() error {
	err := v.Update()
	if err != nil {
		return err
	}
	return v.Render()
}

func (v *Filter) Layout(g *gocui.Gui, minX, minY, maxX, maxY int) error {
	logrus.Tracef("view.Layout(minX: %d, minY: %d, maxX: %d, maxY: %d) %s", minX, minY, maxX, maxY, v.Name())

	label, labelErr := g.SetView(v.Name()+"label", minX, minY, len(v.labelStr), maxY)
	view, viewErr := g.SetView(v.Name(), minX+(len(v.labelStr)-1), minY, maxX, maxY)

	if utils.IsNewView(viewErr, labelErr) {
		err := v.Setup(view, label)
		if err != nil {
			logrus.Error("unable to setup status controller", err)
			return err
		}
	}
	return nil
}

func (v *Filter) RequestedSize(available int) *int {
	return &v.requestedHeight
}
