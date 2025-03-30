package key

import (
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/awesome-gocui/keybinding"
	"github.com/wagoodman/dive/runtime/ui/v1/format"
)

type BindingInfo struct {
	Key        gocui.Key
	Modifier   gocui.Modifier
	Config     Config
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
		if len(info.Config.Keys) == 0 {
			return nil, fmt.Errorf("no keybinding configured for '%s'", info.Display)
		}

		binding, err := newBinding(gui, influence, info.Config.Keys, info.Display, info.OnAction)

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
