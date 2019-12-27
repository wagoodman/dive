package key

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/keybinding"
)

type BindingInfo struct {
	Key        gocui.Key
	Modifier   gocui.Modifier
	ConfigKeys []string
	OnAction   func() error
	IsSelected func() bool
	Display    string
}

type Binding struct {
	key         []keybinding.Key
	displayName string
	selectedFn  func() bool
	actionFn    func() error
}

func GenerateBindings(gui *gocui.Gui, influence string, infos []BindingInfo) ([]*Binding, error) {
	var result = make([]*Binding, 0)
	for _, info := range infos {
		var err error
		var binding *Binding

		if info.ConfigKeys != nil && len(info.ConfigKeys) > 0 {
			binding, err = NewBindingFromConfig(gui, influence, info.ConfigKeys, info.Display, info.OnAction)
		} else {
			binding, err = NewBinding(gui, influence, info.Key, info.Modifier, info.Display, info.OnAction)
		}

		if err != nil {
			return nil, err
		}

		if info.IsSelected != nil {
			binding.RegisterSelectionFn(info.IsSelected)
		}
		if len(info.Display) > 0 {
			result = append(result, binding)
		}
	}
	return result, nil
}

func NewBinding(gui *gocui.Gui, influence string, key gocui.Key, mod gocui.Modifier, displayName string, actionFn func() error) (*Binding, error) {
	return newBinding(gui, influence, []keybinding.Key{{Value: key, Modifier: mod}}, displayName, actionFn)
}

func NewBindingFromConfig(gui *gocui.Gui, influence string, configKeys []string, displayName string, actionFn func() error) (*Binding, error) {
	var parsedKeys []keybinding.Key
	for _, configKey := range configKeys {
		bindStr := viper.GetString(configKey)
		if bindStr == "" {
			logrus.Debugf("skipping keybinding '%s' (no value given)", configKey)
			continue
		}
		logrus.Debugf("parsing keybinding '%s' --> '%s'", configKey, bindStr)

		keys, err := keybinding.ParseAll(bindStr)
		if err != nil {
			return nil, err
		}
		if len(keys) > 0 {
			parsedKeys = keys
			break
		}
	}

	if parsedKeys == nil {
		return nil, fmt.Errorf("could not find configured keybindings for: %+v", configKeys)
	}

	return newBinding(gui, influence, parsedKeys, displayName, actionFn)
}

func newBinding(gui *gocui.Gui, influence string, keys []keybinding.Key, displayName string, actionFn func() error) (*Binding, error) {
	binding := &Binding{
		key:         keys,
		displayName: displayName,
		actionFn:    actionFn,
	}

	for _, key := range keys {
		if err := gui.SetKeybinding(influence, key.Value, key.Modifier, binding.onAction); err != nil {
			return nil, err
		}
	}

	return binding, nil
}

func (binding *Binding) RegisterSelectionFn(selectedFn func() bool) {
	binding.selectedFn = selectedFn
}

func (binding *Binding) onAction(*gocui.Gui, *gocui.View) error {
	if binding.actionFn == nil {
		return fmt.Errorf("no action configured for '%+v'", binding)
	}
	return binding.actionFn()
}

func (binding *Binding) isSelected() bool {
	if binding.selectedFn == nil {
		return false
	}

	return binding.selectedFn()
}

func (binding *Binding) RenderKeyHelp() string {
	return format.RenderHelpKey(binding.key[0].String(), binding.displayName, binding.isSelected())
}
