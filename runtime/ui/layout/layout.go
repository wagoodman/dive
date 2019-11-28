package layout

import "github.com/jroimartin/gocui"

type Layout interface {
	Name() string
	Layout(g *gocui.Gui, minX, minY, maxX, maxY int) error
	RequestedSize(available int) *int
	IsVisible() bool
	OnLayoutChange() error
}
