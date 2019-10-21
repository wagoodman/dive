package layout

import (
	"fmt"
	"github.com/wagoodman/dive/runtime/ui/view"
)

type Vertical struct {
	visible bool
	width int
	elements []View
}

// how does overrun work? which view gets precidence? how does max possible height work?

func NewVerticalLayout() *Vertical {
	return &Vertical{
		visible: true,
		width: view.WidthFull,
		elements: make([]View, 0),
	}
}

func (v Vertical) SetWidth(w int) {
	v.width = w
}

func (v *Vertical) AddView(sub View) error {
	for _, element := range v.elements {
		if element.Name() == sub.Name() {
			return fmt.Errorf("view already added")
		}
	}
	v.elements = append(v.elements, sub)
	return nil
}

func (v *Vertical) Name() string {
	return view.IdentityNone
}

func (v *Vertical) IsVisible() bool {
	return v.visible
}

func (v *Vertical) Height() (height int) {
	for _, element := range v.elements {
		height += element.Height()
		if height == view.HeightFull {
			return view.HeightFull
		}
	}
	return
}

func (v *Vertical) Width() int {
	return v.width
}
