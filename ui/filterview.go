package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

// with special thanks to https://gist.github.com/jroimartin/3b2e943a3811d795e0718b4a95b89bec

type FilterView struct {
	Name      string
	gui       *gocui.Gui
	view      *gocui.View
	header    *gocui.View
	headerStr string
	maxLength int
	hidden    bool
}

type Input struct {
	name      string
	x, y      int
	w         int
	maxLength int
}

func NewFilterView(name string, gui *gocui.Gui) (filterview *FilterView) {
	filterview = new(FilterView)

	// populate main fields
	filterview.Name = name
	filterview.gui = gui
	filterview.headerStr = "Path Filter: "
	filterview.hidden = true

	return filterview
}

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

func (view *FilterView) IsVisible() bool {
	if view == nil {return false}
	return !view.hidden
}

func (view *FilterView) CursorDown() error {
	return nil
}

func (view *FilterView) CursorUp() error {
	return nil
}

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

func (view *FilterView) KeyHelp() string {
	return Formatting.StatusControlNormal("▏Type to filter the file tree ")
}

func (view *FilterView) Update() error {
	return nil
}

func (view *FilterView) Render() error {
	view.gui.Update(func(g *gocui.Gui) error {
		// render the header
		fmt.Fprintln(view.header, Formatting.Header(view.headerStr))

		return nil
	})
	// todo: blerg
	return nil
}
