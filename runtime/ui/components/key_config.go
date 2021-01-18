package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/spf13/viper"
)

// TODO move this to a more appropriate place
type KeyConfig struct{}


type KeyBinding struct{
	*tcell.EventKey
	Display    string
}

type KeyBindingDisplay struct {
	*KeyBinding
	Selected bool
	Hide bool
}

// just print space key as Space
func (kb *KeyBindingDisplay) Name() string {
	if kb.Key() == tcell.KeyRune && kb.Rune() == rune(' ') {
		return "Space"
	}

	return kb.KeyBinding.Name()
}

type keyAction func() bool


func NewKeyBinding(name string, key *tcell.EventKey) KeyBinding {
	return KeyBinding{
		EventKey: key,
		Display: name,
	}
}

func NewKeyBindingDisplay(k tcell.Key, ch rune, modMask tcell.ModMask, name string, selected bool, hide bool) KeyBindingDisplay {
	kb := NewKeyBinding(name, tcell.NewEventKey(k, ch, modMask))
	return KeyBindingDisplay{
		KeyBinding: &kb,
		Selected: selected, 
		Hide: hide,
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
