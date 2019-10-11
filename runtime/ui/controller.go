package ui

import (
	"github.com/jroimartin/gocui"
)

type Renderable interface {
	Update() error
	Render() error
}

// Controller defines the a renderable terminal screen pane.
type Controller interface {
	Renderable
	Setup(*gocui.View, *gocui.View) error
	CursorDown() error
	CursorUp() error
	KeyHelp() string
	IsVisible() bool
}
