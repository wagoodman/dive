package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/components/helpers"
	"github.com/wagoodman/dive/runtime/ui/format"
)

type BoundView interface {
	HasFocus() bool
	GetKeyBindings() []helpers.KeyBindingDisplay
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

func (t *KeyMenuView) GetKeyBindings() []helpers.KeyBindingDisplay {
	logrus.Debug("Getting binding keys from keybinding primitive")
	result := []helpers.KeyBindingDisplay{}
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

	lines := []string{}
	keyBindings := t.GetKeyBindings()
	for idx, binding := range keyBindings {
		if binding.Hide() {
			continue
		}
		displayFormatter := format.StatusControlNormal
		keyBindingFormatter := format.StatusControlNormalBold
		if binding.Selected() {
			displayFormatter = format.StatusControlSelected
			keyBindingFormatter = format.StatusControlSelectedBold
		}
		postfix := "⎹"
		//postfix := "▏"
		if idx == len(keyBindings)-1 {
			postfix = " "
		}
		prefix := " "
		if idx == 0 {
			prefix = ""
		}
		keyBindingContent := keyBindingFormatter(prefix + binding.Name() + " ")
		displayContnet := displayFormatter(binding.Display + postfix)
		lines = append(lines, fmt.Sprintf("%s%s", keyBindingContent, displayContnet))
	}
	joinedLine := strings.Join(lines, "")
	_, w := tview.PrintWithStyle(screen, joinedLine, x, y, width, tview.AlignLeft, tcell.StyleDefault)
	format.PrintLine(screen, format.StatusControlNormal(strings.Repeat(" ", intMax(0, width-w))), x+w, y, width, tview.AlignLeft, tcell.StyleDefault)

}
