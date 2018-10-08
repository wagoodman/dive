package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/lunixbochs/vtclean"
)

type DetailsView struct {
	Name       string
	gui        *gocui.Gui
	view       *gocui.View
	header     *gocui.View
}

func NewStatisticsView(name string, gui *gocui.Gui) (detailsview *DetailsView) {
	detailsview = new(DetailsView)

	// populate main fields
	detailsview.Name = name
	detailsview.gui = gui

	return detailsview
}

func (view *DetailsView) Setup(v *gocui.View, header *gocui.View) error {

	// set view options
	view.view = v
	view.view.Editable = false
	view.view.Wrap = false
	view.view.Highlight = false
	view.view.Frame = false

	view.header = header
	view.header.Editable = false
	view.header.Wrap = false
	view.header.Frame = false

	// set keybindings
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowDown, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorDown() }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowUp, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorUp() }); err != nil {
		return err
	}

	return view.Render()
}

func (view *DetailsView) IsVisible() bool {
	if view == nil {return false}
	return true
}

func (view *DetailsView) Update() error {
	return nil
}

func (view *DetailsView) Render() error {
	view.gui.Update(func(g *gocui.Gui) error {
		// update header
		headerStr := fmt.Sprintf("Image & Layer Details")
		fmt.Fprintln(view.header, Formatting.Header(vtclean.Clean(headerStr, false)))

		// update contents
		view.view.Clear()
		fmt.Fprintln(view.view, "TBD")
		return nil
	})
	return nil
}

func (view *DetailsView) CursorDown() error {
	return CursorDown(view.gui, view.view)
}

func (view *DetailsView) CursorUp() error {
	return CursorUp(view.gui, view.view)
}


func (view *DetailsView) KeyHelp() string {
	return "TBD"
	// return  renderStatusOption("^L","Layer changes", view.CompareMode == CompareLayer) +
	// 		renderStatusOption("^A","All changes", view.CompareMode == CompareAll)
}
