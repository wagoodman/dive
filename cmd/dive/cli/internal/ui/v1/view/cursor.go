package view

import (
	"errors"

	"github.com/awesome-gocui/gocui"
)

// CursorDown moves the cursor down in the currently selected gocui pane, scrolling the screen as needed.
func CursorDown(g *gocui.Gui, v *gocui.View) error {
	return CursorStep(g, v, 1)
}

// CursorUp moves the cursor up in the currently selected gocui pane, scrolling the screen as needed.
func CursorUp(g *gocui.Gui, v *gocui.View) error {
	return CursorStep(g, v, -1)
}

// Moves the cursor the given step distance, setting the origin to the new cursor line
func CursorStep(g *gocui.Gui, v *gocui.View, step int) error {
	cx, cy := v.Cursor()

	// if there isn't a next line
	line, err := v.Line(cy + step)
	if err != nil {
		return err
	}
	if len(line) == 0 {
		return errors.New("unable to move the cursor, empty line")
	}
	if err := v.SetCursor(cx, cy+step); err != nil {
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+step); err != nil {
			return err
		}
	}
	return nil
}
