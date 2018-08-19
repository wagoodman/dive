package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

type StatusView struct {
	Name string
	gui  *gocui.Gui
	view *gocui.View
}

func NewStatusView(name string, gui *gocui.Gui) (statusview *StatusView) {
	statusview = new(StatusView)

	// populate main fields
	statusview.Name = name
	statusview.gui = gui

	return statusview
}

func (view *StatusView) Setup(v *gocui.View, header *gocui.View) error {

	// set view options
	view.view = v
	view.view.Frame = false
	view.view.BgColor = gocui.ColorDefault + gocui.AttrReverse

	view.Render()

	return nil
}

func (view *StatusView) IsVisible() bool {
	if view == nil {return false}
	return true
}

func (view *StatusView) CursorDown() error {
	return nil
}

func (view *StatusView) CursorUp() error {
	return nil
}

func (view *StatusView) KeyHelp() string {
	return Formatting.Control("[^C]") + ": Quit " +
		Formatting.Control("[^Space]") + ": Switch View " +
     	Formatting.Control("[^/]") + ": Filter files"
}

func (view *StatusView) Update() error {
	return nil
}

func (view *StatusView) Render() error {
	view.gui.Update(func(g *gocui.Gui) error {
		view.view.Clear()
		fmt.Fprintln(view.view, view.KeyHelp()+" | "+Views.lookup[view.gui.CurrentView().Name()].KeyHelp())

		return nil
	})
	// todo: blerg
	return nil
}
