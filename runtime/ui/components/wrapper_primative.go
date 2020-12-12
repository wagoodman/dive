package components

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

// This was pretty helpful: https://github.com/rivo/tview/wiki/Primitives

type wrapable interface {
	getBox() *tview.Box
	getDraw() drawFn
	getInputWrapper() inputFn
}

type Wrapper struct {
	*tview.Box
	flex             *tview.Flex
	titleTextView    *tview.TextView
	subtitleTextView *tview.TextView
	inner            wrapable
	titleRightBox    *tview.Box
	title            string
	subtitle         string
}

type drawFn func(screen tcell.Screen)
type inputFn func() func(event *tcell.EventKey, setFocus func(p tview.Primitive))

func NewWrapper(title, subtitle string, inner wrapable) *Wrapper {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	w := &Wrapper{
		Box:              flex.Box,
		flex:             flex,
		title:            title,
		subtitle:         subtitle,
		titleTextView:    tview.NewTextView(),
		subtitleTextView: tview.NewTextView().SetText(subtitle),
		inner:            inner,
	}
	w.setTitle(w.inner.getBox().HasFocus())
	return w
}

func (b *Wrapper) Setup() *Wrapper {
	if b.title != "" {
		innerflex := tview.NewFlex().SetDirection(tview.FlexColumn)
		titleRightBox := tview.NewBox()
		b.titleRightBox = titleRightBox

		innerflex.AddItem(b.titleTextView.SetWrap(false).SetBorder(false), len(b.title)+10, 1, false)
		innerflex.AddItem(titleRightBox, 0, 1, false)
		b.flex.AddItem(innerflex, 1, 1, false)
	}

	if b.subtitle != "" {
		b.flex.AddItem(b.subtitleTextView.SetText(b.subtitle).SetWrap(false).SetBorder(false), 1, 1, false)
	}

	b.flex.AddItem(b.inner.getBox().SetBorder(false), 0, 1, true)
	return b
}

func (b *Wrapper) Draw(screen tcell.Screen) {
	b.flex.Draw(screen)

	if b.title != "" {
		b.titleTextView.Draw(screen)
		x, y, width, _ := b.titleRightBox.GetRect()
		for cx := x; cx < x+width; cx++ {
			// TODO: swap for bold when focused (tview.BoxDrawingsHeavyHorizontal)
			screen.SetContent(cx, y, tview.BoxDrawingsLightHorizontal, nil, tcell.StyleDefault)
		}
	}
	if b.subtitle != "" {
		b.subtitleTextView.Draw(screen)
	}
	b.inner.getDraw()(screen)
}

func (b *Wrapper) Focus(delegate func(p tview.Primitive)) {
	if b.title != "" {
		b.setTitle(true)
	}
	b.inner.getBox().Focus(delegate)
}

func (b *Wrapper) Blur() {
	if b.title != "" {
		b.setTitle(false)
	}
	b.inner.getBox().Blur()
}

func (b *Wrapper) HasFocus() bool {
	return b.inner.getBox().HasFocus()
}

func (b *Wrapper) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return b.inner.getInputWrapper()()
}

func (b *Wrapper) setTitle(hasFocus bool) {
	prefix := ""
	if hasFocus {
		prefix = "● "
	}
	// TODO: swap for bold when focused (tview.BoxDrawingsHeavyHorizontal)
	postfix := strings.Repeat(string(tview.BoxDrawingsLightHorizontal), 10)
	title := fmt.Sprintf("│ %s%s ├%s", prefix, b.title, postfix)

	if hasFocus {
		// TODO: add bold wrapper
	}

	b.titleTextView.SetText(title)
}
