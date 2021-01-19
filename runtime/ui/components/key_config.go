package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/spf13/viper"
)

// TODO move this to a more appropriate place
type KeyConfig struct{}

type KeyBinding struct {
	*tcell.EventKey
	Display string
}

type KeyBindingDisplay struct {
	*KeyBinding
	Selected bool
	Hide     bool
}

func (kb *KeyBindingDisplay) Name() string {
	s := ""
	m := []string{}
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
			if kb.Rune() == rune(' ') {
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
			s = s[5:]
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

func NewKeyBindingDisplay(k tcell.Key, ch rune, modMask tcell.ModMask, name string, selected bool, hide bool) KeyBindingDisplay {
	kb := NewKeyBinding(name, tcell.NewEventKey(k, ch, modMask))
	return KeyBindingDisplay{
		KeyBinding: &kb,
		Selected:   selected,
		Hide:       hide,
	}
}

func (k *KeyBinding) Match(event *tcell.EventKey) bool {
	if k.Key() == tcell.KeyRune {
		return k.Rune() == event.Rune() && (k.Modifiers() == event.Modifiers())
	}

	return k.Key() == event.Key()
}

type MissingConfigError struct {
	Field string
}

func NewMissingConfigErr(field string) MissingConfigError {
	return MissingConfigError{
		Field: field,
	}
}

func (e MissingConfigError) Error() string {
	return fmt.Sprintf("error configuration %s: not found", e.Field)
}

func NewKeyConfig() *KeyConfig {
	return &KeyConfig{}
}

func (k *KeyConfig) GetKeyBinding(key string) (result KeyBinding, err error) {
	err = viper.UnmarshalKey(key, &result)
	if err != nil {
		return KeyBinding{}, err
	}
	return result, err
}
