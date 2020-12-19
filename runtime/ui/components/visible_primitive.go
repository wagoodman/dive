package components

import (
	"github.com/rivo/tview"
)

type Visiblility interface {
	Visible() bool
}

type VisiblePrimitive interface {
	tview.Primitive
	Visiblility
}

type VisibleFunc func(tview.Primitive) bool

func Always(alwaysVal bool) VisibleFunc {
	return func (_ tview.Primitive) bool {
		return alwaysVal
	}
}

func MinHeightVisibility(minHeight int) VisibleFunc {
	return func(p tview.Primitive) bool {
		_, _, _, height := p.GetRect()
		return height >= minHeight
	}
}

// How can we actually implement this????
// Either we have to do one of the following
// 1) we want to use particular and specific methods on an item
//    - we Have to make VisibleFunc methods know what their base class is ( or at least have a larger interface )
// 2) How can we make this configurable
// 3) make this an implementaion detail of each struct that conforms to this interface (this seems like the best idea)
