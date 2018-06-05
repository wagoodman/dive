package ui

import 	"github.com/jroimartin/gocui"

type View interface {
	Setup(*gocui.View) error
	CursorDown() error
	CursorUp() error
	Render() error
}
