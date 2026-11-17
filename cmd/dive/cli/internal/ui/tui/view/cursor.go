package view

import (
	"strings"

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

// CursorStep moves the cursor the given number of lines (clamped to the content), scrolling the view when the
// cursor would leave it.
func CursorStep(_ *gocui.Gui, v *gocui.View, step int) error {
	// note: the cursor row is relative to the view (it's the highlighted row on screen) while the origin indexes
	// the wrapped buffer lines, so work in absolute line positions (origin + cursor)
	lines := v.ViewBufferLines()
	// the trailing newline(s) of the content leave blank lines that shouldn't be selectable
	last := len(lines) - 1
	for last > 0 && strings.TrimSpace(lines[last]) == "" {
		last--
	}
	_, height := v.Size()
	if last < 0 || height <= 0 {
		return nil
	}

	cx, cy := v.Cursor()
	ox, oy := v.Origin()
	target := max(0, min(oy+cy+step, last))

	oy = min(oy, target)
	oy = max(oy, target-height+1)
	if err := v.SetOrigin(ox, oy); err != nil {
		return err
	}
	return v.SetCursorUnrestricted(cx, target-oy)
}
