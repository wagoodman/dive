package constructors

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/runtime/ui/components/helpers"
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

func (k *KeyConfig) GetKeyBinding(key string) (helpers.KeyBinding, error) {
	name, ok := DisplayNames[key]
	if !ok {
		return helpers.KeyBinding{}, fmt.Errorf("no name for binding %q found", key)
	}
	keyName := viper.GetString(key)
	mod, tKey, ch, err := cbind.Decode(keyName)
	if err != nil {
		return helpers.KeyBinding{}, fmt.Errorf("unable to create binding from dive.config file: %q", err)
	}
	fmt.Printf("creating key event for %s\n", key)
	fmt.Printf("mod %d, key %d, ch %s\n", mod, tKey, string(ch))
	return helpers.NewKeyBinding(name, tcell.NewEventKey(tKey, ch, mod)), nil
}
