package helpers

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type KeyBinding struct {
	*tcell.EventKey
	Display string
}

type KeyBindingDisplay struct {
	*KeyBinding
	Selected func() bool
	Hide     func() bool
}

func (kb *KeyBindingDisplay) Name() string {
	var s string
	var m []string
	kMod := kb.Modifiers()
	if kMod&tcell.ModShift != 0 {
		m = append(m, "Shift")
	}
	if kMod&tcell.ModAlt != 0 {
		m = append(m, "Alt")
	}
	if kMod&tcell.ModMeta != 0 {
		m = append(m, "Meta")
	}
	if kMod&tcell.ModCtrl != 0 {
		m = append(m, "^")
	}

	ok := false
	key := kb.Key()
	if s, ok = tcell.KeyNames[key]; !ok {
		if key == tcell.KeyRune {
			if kb.Rune() == ' ' {
				s = "Space"
			} else {
				s = string(kb.Rune())
			}
		} else {
			s = fmt.Sprintf("Key[%d,%d]", key, int(kb.Rune()))
		}
	}
	if len(m) != 0 {
		if kMod&tcell.ModCtrl != 0 && strings.HasPrefix(s, "Ctrl-") {
			s = strings.TrimPrefix(s, "Ctrl-")
		}
		return fmt.Sprintf("%s%s", strings.Join(m, ""), s)
	}
	return s
}

func NewKeyBinding(name string, key *tcell.EventKey) KeyBinding {
	return KeyBinding{
		EventKey: key,
		Display:  name,
	}
}

func (k *KeyBinding) Match(event *tcell.EventKey) bool {
	if k.Key() == tcell.KeyRune {
		return k.Rune() == event.Rune() && (k.Modifiers() == event.Modifiers())
	}

	return k.Key() == event.Key()
}
