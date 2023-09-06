package view

import (
	"github.com/awesome-gocui/gocui"
)

// CursorDown moves the cursor down in the currently selected gocui pane, scrolling the screen as needed.
func CursorDown(v *gocui.View, step uint) error {
	v.MoveCursor(0, int(step))
	return nil
}

// CursorUp moves the cursor up in the currently selected gocui pane, scrolling the screen as needed.
func CursorUp(v *gocui.View, step uint) error {
	v.MoveCursor(0, int(-step))
	return nil
}
