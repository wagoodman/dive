package ui

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
)

// with special thanks to https://gist.github.com/jroimartin/3b2e943a3811d795e0718b4a95b89bec

type CommandView struct {
	Name      string
	gui       *gocui.Gui
	view      *gocui.View
	maxLength int
}

type Input struct {
	name      string
	x, y      int
	w         int
	maxLength int
}

func NewCommandView(name string, gui *gocui.Gui) (commandview *CommandView) {
	commandview = new(CommandView)

	// populate main fields
	commandview.Name = name
	commandview.gui = gui

	return commandview
}

func (view *CommandView) Setup(v *gocui.View) error {

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

func (view *CommandView) CursorDown() error {
	return nil
}

func (view *CommandView) CursorUp() error {
	return nil
}

func (i *CommandView) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	cx, _ := v.Cursor()
	ox, _ := v.Origin()
	limit := ox+cx+1 > i.maxLength
	switch {
	case ch != 0 && mod == 0 && !limit:
		v.EditWrite(ch)
	case key == gocui.KeySpace && !limit:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	}
}

func (view *CommandView) KeyHelp() string {
	control := color.New(color.Bold).SprintFunc()
	return control("[^C]") + ": Quit " +
		control("[^Space]") + ": Switch View "

}

func (view *CommandView) Render() error {
	view.gui.Update(func(g *gocui.Gui) error {
		view.view.Clear()
		fmt.Fprintln(view.view, "Filter: ")

		return nil
	})
	// todo: blerg
	return nil
}
