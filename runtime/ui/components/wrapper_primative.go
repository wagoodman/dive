package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wagoodman/dive/runtime/ui/components/helpers"
	"github.com/wagoodman/dive/runtime/ui/format"
)

// This was pretty helpful: https://github.com/rivo/tview/wiki/Primitives

type wrappable interface {
	getBox() *tview.Box
	getDraw() drawFn
	getInputWrapper() inputFn
	GetKeyBindings() []helpers.KeyBindingDisplay
}

type Wrapper struct {
	*tview.Box
	flex             *tview.Flex
	titleTextView    *tview.TextView
	subtitleTextView *tview.TextView
	inner            wrappable
	titleRightBox    *tview.Box
	title            string
	subtitle         string
	visible          VisibleFunc
	getKeyBindings   func() []helpers.KeyBindingDisplay
}

type drawFn func(screen tcell.Screen)
type inputFn func() func(event *tcell.EventKey, setFocus func(p tview.Primitive))

func NewWrapper(title, subtitle string, inner wrappable) *Wrapper {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	w := &Wrapper{
		Box:              flex.Box,
		flex:             flex,
		title:            title,
		subtitle:         subtitle,
		titleTextView:    tview.NewTextView().SetDynamicColors(true),
		subtitleTextView: tview.NewTextView().SetText(subtitle),
		inner:            inner,
		visible:          Always(true),
		getKeyBindings:   inner.GetKeyBindings,
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
		boarderRune := tview.BoxDrawingsLightHorizontal
		if b.HasFocus() {
			boarderRune = tview.BoxDrawingsHeavyHorizontal
		}
		for cx := x; cx < x+width; cx++ {
			screen.SetContent(cx, y, boarderRune, nil, tcell.StyleDefault)
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
	postfix := strings.Repeat(string(tview.BoxDrawingsLightHorizontal), 10)
	if hasFocus {
		prefix = "● "
		postfix = strings.Repeat(string(tview.BoxDrawingsHeavyHorizontal), 10)
	}
	content := format.Header(fmt.Sprintf("%s%s", prefix, b.title))
	title := fmt.Sprintf("│ %s ├%s", content, postfix)

	if hasFocus {
		content = format.Header(fmt.Sprintf("%s%s", prefix, b.title))
		title = fmt.Sprintf("┃ %s ┣%s", content, postfix)
	}

	b.titleTextView.SetText(title)
}

func (b *Wrapper) Visible() bool {
	return b.visible(b)
}

func (b *Wrapper) SetVisibility(visibleFunc VisibleFunc) *Wrapper {
	b.visible = visibleFunc
	return b
}

func (b *Wrapper) GetKeyBindings() []helpers.KeyBindingDisplay {
	return b.getKeyBindings()
}
