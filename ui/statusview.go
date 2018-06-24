package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/fatih/color"
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

	// set keybindings
	// if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowDown, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorDown() }); err != nil {
	// 	return err
	// }
	// if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowUp, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorUp() }); err != nil {
	// 	return err
	// }

	view.Render()

	return nil
}


func (view *StatusView) CursorDown() error {
	return nil
}

func (view *StatusView) CursorUp() error {
	return nil
}

func (view *StatusView) KeyHelp() string {
	control := color.New(color.Bold).SprintFunc()
	return  control("[^C]") + ": Quit " +
		control("[^Space]") + ": Switch View "

}

func (view *StatusView) Render() error {
	view.gui.Update(func(g *gocui.Gui) error {
		view.view.Clear()
		fmt.Fprintln(view.view, view.KeyHelp() + " | " + Views.lookup[view.gui.CurrentView().Name()].KeyHelp())

		return nil
	})
	// todo: blerg
	return nil
}
