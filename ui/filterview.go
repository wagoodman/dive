package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

// DetailsView holds the UI objects and data models for populating the bottom row. Specifically the pane that
// allows the user to filter the file tree by path.
type FilterView struct {
	Name      string
	gui       *gocui.Gui
	view      *gocui.View
	header    *gocui.View
	headerStr string
	maxLength int
	hidden    bool
}

// NewFilterView creates a new view object attached the the global [gocui] screen object.
func NewFilterView(name string, gui *gocui.Gui) (filterView *FilterView) {
	filterView = new(FilterView)

	// populate main fields
	filterView.Name = name
	filterView.gui = gui
	filterView.headerStr = "Path Filter: "
	filterView.hidden = true

	return filterView
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (view *FilterView) Setup(v *gocui.View, header *gocui.View) error {

	// set view options
	view.view = v
	view.maxLength = 200
	view.view.Frame = false
	view.view.BgColor = gocui.AttrReverse
	view.view.Editable = true
	view.view.Editor = view

	view.header = header
	view.header.BgColor = gocui.AttrReverse
	view.header.Editable = false
	view.header.Wrap = false
	view.header.Frame = false

	view.Render()

	return nil
}

// IsVisible indicates if the filter view pane is currently initialized
func (view *FilterView) IsVisible() bool {
	if view == nil {
		return false
	}
	return !view.hidden
}

// CursorDown moves the cursor down in the filter pane (currently indicates nothing).
func (view *FilterView) CursorDown() error {
	return nil
}

// CursorUp moves the cursor up in the filter pane (currently indicates nothing).
func (view *FilterView) CursorUp() error {
	return nil
}

// Edit intercepts the key press events in the filer view to update the file view in real time.
func (view *FilterView) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if !view.IsVisible() {
		return
	}

	cx, _ := v.Cursor()
	ox, _ := v.Origin()
	limit := ox+cx+1 > view.maxLength
	switch {
	case ch != 0 && mod == 0 && !limit:
		v.EditWrite(ch)
	case key == gocui.KeySpace && !limit:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	}
	if Views.Tree != nil {
		Views.Tree.Update()
		Views.Tree.Render()
	}
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (view *FilterView) Update() error {
	return nil
}

// Render flushes the state objects to the screen. Currently this is the users path filter input.
func (view *FilterView) Render() error {
	view.gui.Update(func(g *gocui.Gui) error {
		// render the header
		fmt.Fprintln(view.header, Formatting.Header(view.headerStr))

		return nil
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected.
func (view *FilterView) KeyHelp() string {
	return Formatting.StatusControlNormal("▏Type to filter the file tree ")
}
