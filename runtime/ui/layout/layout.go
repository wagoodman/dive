package layout

import "github.com/jroimartin/gocui"

type Layout interface {
	Name() string
	Layout(g *gocui.Gui, minX, minY, maxX, maxY int, hasResized bool) error
	RequestedSize(available int) *int
	IsVisible() bool
}
