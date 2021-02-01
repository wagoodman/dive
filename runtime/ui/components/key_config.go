package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cbind"
)

// TODO move key constants out to their own file
var DisplayNames = map[string]string{
	"keybinding.quit":                       "Quit",
	"keybinding.toggle-view":                "Switch View",
	"keybinding.filter-files":               "Find",
	"keybinding.compare-all":                "Compare All",
	"keybinding.compare-layer":              "Compare Layer",
	"keybinding.toggle-collapse-dir":        "Collapse",
	"keybinding.toggle-collapse-all-dir":    "Collapse All",
	"keybinding.toggle-filetree-attributes": "Attributes",
	"keybinding.toggle-added-files":         "Added",
	"keybinding.toggle-removed-files":       "Removed",
	"keybinding.toggle-modified-files":      "Modified",
	"keybinding.toggle-unmodified-files":    "Unmodified",
	"keybinding.page-up":                    "Pg Up",
	"keybinding.page-down":                  "Pg Down",
}

// TODO move this to a more appropriate place
type KeyConfig struct{}

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

func (k *KeyConfig) GetKeyBinding(key string) (KeyBinding, error) {
	name, ok := DisplayNames[key]
	if !ok {
		return KeyBinding{}, fmt.Errorf("no name for binding %q found", key)
	}
	keyName := viper.GetString(key)
	mod, tKey, ch, err := cbind.Decode(keyName)
	if err != nil {
		return KeyBinding{}, fmt.Errorf("unable to create binding from dive.config file: %q", err)
	}
	fmt.Printf("creating key event for %s\n", key)
	fmt.Printf("mod %d, key %d, ch %s\n", mod, tKey, string(ch))
	return NewKeyBinding(name, tcell.NewEventKey(tKey, ch, mod)), nil
}
