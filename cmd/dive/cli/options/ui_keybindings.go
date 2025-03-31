package options

import (
	"fmt"
	"github.com/anchore/clio"
	"github.com/wagoodman/dive/runtime/ui/v1/key"
	"reflect"
)

var _ interface {
	clio.FieldDescriber
} = (*UIKeybindings)(nil)

// UIKeybindings provides configuration for all keyboard shortcuts
type UIKeybindings struct {
	key.Bindings `yaml:",inline" mapstructure:",squash"`
}

func DefaultUIKeybinding() UIKeybindings {
	return UIKeybindings{
		Bindings: key.DefaultBindings(),
	}
}

func (c *UIKeybindings) DescribeFields(descriptions clio.FieldDescriptionSet) {
	// global keybindings
	descriptions.Add(&c.Global.Quit, "quit the application (global)")
	descriptions.Add(&c.Global.ToggleView, "toggle between different views (global)")
	descriptions.Add(&c.Global.FilterFiles, "filter files by name (global)")
	descriptions.Add(&c.Global.CloseFilterFiles, "close file filtering (global)")

	// navigation keybindings
	descriptions.Add(&c.Navigation.Up, "move cursor up (global)")
	descriptions.Add(&c.Navigation.Down, "move cursor down (global)")
	descriptions.Add(&c.Navigation.Left, "move cursor left (global)")
	descriptions.Add(&c.Navigation.Right, "move cursor right (global)")
	descriptions.Add(&c.Navigation.PageUp, "scroll page up (file view)")
	descriptions.Add(&c.Navigation.PageDown, "scroll page down (file view)")

	// layer view keybindings
	descriptions.Add(&c.Layer.CompareAll, "compare all layers (layer view)")
	descriptions.Add(&c.Layer.CompareLayer, "compare specific layer (layer view)")

	// file view keybindings
	descriptions.Add(&c.Filetree.ToggleCollapseDir, "toggle directory collapse (file view)")
	descriptions.Add(&c.Filetree.ToggleCollapseAllDir, "toggle collapse all directories (file view)")
	descriptions.Add(&c.Filetree.ToggleAddedFiles, "toggle visibility of added files (file view)")
	descriptions.Add(&c.Filetree.ToggleRemovedFiles, "toggle visibility of removed files (file view)")
	descriptions.Add(&c.Filetree.ToggleModifiedFiles, "toggle visibility of modified files (file view)")
	descriptions.Add(&c.Filetree.ToggleUnmodifiedFiles, "toggle visibility of unmodified files (file view)")
	descriptions.Add(&c.Filetree.ToggleTreeAttributes, "toggle display of file attributes (file view)")
	descriptions.Add(&c.Filetree.ToggleSortOrder, "toggle sort order (file view)")
	descriptions.Add(&c.Filetree.ExtractFile, "extract file contents (file view)")
}

func (c *UIKeybindings) PostLoad() error {
	return recursivelySetupConfigs(reflect.ValueOf(&c.Bindings).Elem())
}

// recursivelySetupConfigs traverses struct fields and calls Setup() on any key.Config fields
func recursivelySetupConfigs(val reflect.Value) error {
	typ := val.Type()

	if typ.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < typ.NumField(); i++ {
		field := val.Field(i)

		if field.Type() == reflect.TypeOf(key.Config{}) {
			if field.CanAddr() {
				configPtr := field.Addr().Interface().(*key.Config)

				if err := configPtr.Setup(); err != nil {
					fieldName := typ.Field(i).Name
					return fmt.Errorf("failed to set up key binding for %s: %w", fieldName, err)
				}
			}
		} else if field.Kind() == reflect.Struct {
			if err := recursivelySetupConfigs(field); err != nil {
				return err
			}
		}
	}

	return nil
}
