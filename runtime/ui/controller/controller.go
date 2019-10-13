package controller

import (
	"github.com/jroimartin/gocui"
)

// Controller defines the a renderable terminal screen pane.
type Controller interface {
	Update() error
	Render() error
	Setup(*gocui.View, *gocui.View) error
	CursorDown() error
	CursorUp() error
	KeyHelp() string
	IsVisible() bool
}
