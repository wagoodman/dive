package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/format"
)

type BoundView interface {
	HasFocus() bool
	GetKeyBindings() []KeyBindingDisplay
}

type KeyMenuView struct {
	*tview.TextView
	boundList []BoundView
}

func NewKeyMenuView() *KeyMenuView {
	return &KeyMenuView{
		TextView:  tview.NewTextView(),
		boundList: []BoundView{},
	}
}

func (t *KeyMenuView) AddBoundViews(b ...BoundView) *KeyMenuView {
	t.boundList = append(t.boundList, b...)
	return t
}

func (t *KeyMenuView) RemoveViews(b ...BoundView) *KeyMenuView {
	newBoundList := []BoundView{}
	boundSet := map[BoundView]interface{}{}
	for _, v := range b {
		boundSet[v] = true
	}

	for _, bound := range t.boundList {
		if _, ok := boundSet[bound]; !ok {
			newBoundList = append(newBoundList, bound)
		}
	}

	t.boundList = newBoundList
	return t
}

func (t *KeyMenuView) GetKeyBindings() []KeyBindingDisplay {
	logrus.Debug("Getting binding keys from keybinding primitive")
	result := []KeyBindingDisplay{}
	for _, view := range t.boundList {
		if view.HasFocus() {
			result = append(result, view.GetKeyBindings()...)
		}
	}

	return result
}

func (t *KeyMenuView) Draw(screen tcell.Screen) {
	t.Box.Draw(screen)
	x, y, width, _ := t.Box.GetInnerRect()
	// TODO: add logic to highlight selected options

	line := []string{}
	for _, keyBinding := range t.GetKeyBindings() {
		if keyBinding.Hide {
			continue
		}
		formatter := format.StatusControlNormal
		if keyBinding.Selected {
			formatter = format.StatusControlSelected
		}
		line = append(line, formatter(fmt.Sprintf("%s (%s)", keyBinding.Display, keyBinding.Name())))
	}

	format.PrintLine(screen, strings.Join(line, format.StatusControlNormal(" ▏")), x, y, width, tview.AlignLeft, tcell.StyleDefault)

}

// for wrappers
func (t *KeyMenuView) getBox() *tview.Box {
	return t.Box
}

func (t *KeyMenuView) getDraw() drawFn {
	return t.Draw
}
