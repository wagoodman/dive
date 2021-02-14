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


