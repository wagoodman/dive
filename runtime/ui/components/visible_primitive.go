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

type VisibleWrapper struct {
	tview.Primitive
	visible VisibleFunc
}

func (v *VisibleWrapper) Visible() bool {
	return v.visible(v)
}

func (v *VisibleWrapper) SetVisibility(visibleFunc VisibleFunc) *VisibleWrapper {
	v.visible = visibleFunc
	return v
}

func NewVisibleWrapper(p tview.Primitive) *VisibleWrapper {
	return &VisibleWrapper{
		Primitive: p,
		visible:   AlwaysVisible,
	}
}

func AlwaysVisible(_ tview.Primitive) bool {
	return true
}
