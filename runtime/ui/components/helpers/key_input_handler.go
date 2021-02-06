package helpers

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type InputHandleFunc func(event *tcell.EventKey, setFocus func(p tview.Primitive))

// TODO factor out KeyInputHandler and related structs into a separate file
type KeyInputHandler struct {
	Order      []KeyBindingDisplay
	HandlerMap map[*tcell.EventKey]func()
}

func NewKeyInputHandler() *KeyInputHandler {
	return &KeyInputHandler{
		Order:      []KeyBindingDisplay{},
		HandlerMap: map[*tcell.EventKey]func(){},
	}
}

func (k *KeyInputHandler) AddBinding(binding KeyBindingDisplay, f func()) *KeyInputHandler {
	k.Order = append(k.Order, binding)
	k.HandlerMap[binding.EventKey] = f

	return k
}

func (k *KeyInputHandler) Handle() InputHandleFunc {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		for _, m := range k.Order {
			if m.Match(event) {
				k.HandlerMap[m.EventKey]()
			}
		}
	}
}
