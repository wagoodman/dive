package components

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type LayerList struct {
	*tview.Box
	subtitle string
	// TODO make me an interface
	layers       []string
	cmpIndex int
	changed func(index int, mainText string, shortcut rune)
	selectedBackgroundColor tcell.Color
}

func NewLayerList(options []string) *LayerList {
	return &LayerList{
		Box: tview.NewBox(),
		layers: options,
		cmpIndex: 0,
	}
}

func (ll *LayerList) SetSubtitle(subtitle string) *LayerList {
	ll.subtitle = subtitle
	return ll
}

func (ll *LayerList) GetSubtitle() string {
	return ll.subtitle
}

func(ll *LayerList) GetInnerRect() (int,int,int,int) {
	x,y,width,height := ll.Box.GetInnerRect()
	if ll.subtitle != "" {
		return x, y + 1, width, height - 1
	}
	return x,y,width,height
}

func (ll *LayerList) drawSubtitle(screen tcell.Screen) {
	x,y,width,height := ll.Box.GetInnerRect()
	if height > 1 {
		line := fmt.Sprintf("[white]%s", ll.subtitle)
		tview.Print(screen, line, x, y, width, tview.AlignLeft, tcell.ColorDefault)
	}
}

func (ll *LayerList) Draw(screen tcell.Screen) {
	ll.Box.Draw(screen)
	ll.drawSubtitle(screen)
	x, y, width, height := ll.GetInnerRect()

	cmpString := "  "
	for yIndex, layer := range ll.layers {
		if yIndex > height {
			break
		}
		var cmpColor tcell.Color
		switch {
		case yIndex == ll.cmpIndex:
			cmpColor = tcell.ColorRed
		case yIndex < ll.cmpIndex:
			cmpColor = tcell.ColorBlue
		default:
			cmpColor = tcell.ColorDefault
		}
		line := fmt.Sprintf("%s[white] %s", cmpString, layer)
		tview.Print(screen, line, x, y+yIndex, width, tview.AlignLeft, cmpColor)
		for xIndex := 0; xIndex < width; xIndex++ {
			m, c, style, _ := screen.GetContent(x+xIndex, y+yIndex)
			fg, bg, _ := style.Decompose()
			style = style.Background(fg).Foreground(bg)
			switch {
			case yIndex == ll.cmpIndex:
				screen.SetContent(x+xIndex, y+yIndex, m, c, style)
				screen.SetContent(x+xIndex, y+yIndex, m, c, style)
			case yIndex < ll.cmpIndex && xIndex < len(cmpString):
				screen.SetContent(x+xIndex, y+yIndex, m, c, style)
				screen.SetContent(x+xIndex, y+yIndex, m, c, style)
			default:
				break
			}
		}
	}
}

func (ll *LayerList) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return ll.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp:
			ll.cmpIndex--
			if ll.cmpIndex < 0 {
				ll.cmpIndex = 0
			}
		case tcell.KeyDown:
			ll.cmpIndex++
			if ll.cmpIndex >= len(ll.layers) {
				ll.cmpIndex = len(ll.layers) - 1
			}
		}
		ll.changed(ll.cmpIndex, ll.layers[ll.cmpIndex], event.Rune())
	})
}

func (ll *LayerList) InsertItem(index int, value string) *LayerList {
	if index < 0 {
		ll.layers = append(ll.layers, value)
		return ll
	}
	ll.layers = append(ll.layers[:index+1], ll.layers[index:]...)
	ll.layers[index] = value
	return ll
}

func (ll *LayerList) AddItem(mainText string) *LayerList {
	ll.InsertItem(-1, mainText)
	return ll
}

func (ll *LayerList) Focus(delegate func(p tview.Primitive)) {
	ll.Box.Focus(delegate)
}

func (ll *LayerList) HasFocus() bool {
	return ll.Box.HasFocus()
}

func (ll *LayerList) SetChangedFunc(handler func(index int, mainText string, shortcut rune)) *LayerList {
	ll.changed = handler
	return ll
}