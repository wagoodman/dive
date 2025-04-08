package options

import (
	"github.com/anchore/clio"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/key"
	"reflect"
)

var _ interface {
	clio.FieldDescriber
} = (*UIKeybindings)(nil)

// UIKeybindings provides configuration for all keyboard shortcuts
type UIKeybindings struct {
	Global     GlobalBindings     `yaml:",inline" mapstructure:",squash"`
	Navigation NavigationBindings `yaml:",inline" mapstructure:",squash"`
	Layer      LayerBindings      `yaml:",inline" mapstructure:",squash"`
	Filetree   FiletreeBindings   `yaml:",inline" mapstructure:",squash"`

	Config key.Bindings `yaml:"-" mapstructure:"-"`
}

type GlobalBindings struct {
	Quit             string `yaml:"quit" mapstructure:"quit"`
	ToggleView       string `yaml:"toggle-view" mapstructure:"toggle-view"`
	FilterFiles      string `yaml:"filter-files" mapstructure:"filter-files"`
	CloseFilterFiles string `yaml:"close-filter-files" mapstructure:"close-filter-files"`
}

type NavigationBindings struct {
	Up       string `yaml:"up" mapstructure:"up"`
	Down     string `yaml:"down" mapstructure:"down"`
	Left     string `yaml:"left" mapstructure:"left"`
	Right    string `yaml:"right" mapstructure:"right"`
	PageUp   string `yaml:"page-up" mapstructure:"page-up"`
	PageDown string `yaml:"page-down" mapstructure:"page-down"`
}

type LayerBindings struct {
	CompareAll   string `yaml:"compare-all" mapstructure:"compare-all"`
	CompareLayer string `yaml:"compare-layer" mapstructure:"compare-layer"`
}

type FiletreeBindings struct {
	ToggleCollapseDir     string `yaml:"toggle-collapse-dir" mapstructure:"toggle-collapse-dir"`
	ToggleCollapseAllDir  string `yaml:"toggle-collapse-all-dir" mapstructure:"toggle-collapse-all-dir"`
	ToggleAddedFiles      string `yaml:"toggle-added-files" mapstructure:"toggle-added-files"`
	ToggleRemovedFiles    string `yaml:"toggle-removed-files" mapstructure:"toggle-removed-files"`
	ToggleModifiedFiles   string `yaml:"toggle-modified-files" mapstructure:"toggle-modified-files"`
	ToggleUnmodifiedFiles string `yaml:"toggle-unmodified-files" mapstructure:"toggle-unmodified-files"`
	ToggleTreeAttributes  string `yaml:"toggle-filetree-attributes" mapstructure:"toggle-filetree-attributes"`
	ToggleSortOrder       string `yaml:"toggle-sort-order" mapstructure:"toggle-sort-order"`
	ToggleWrapTree        string `yaml:"toggle-wrap-tree" mapstructure:"toggle-wrap-tree"`
	ExtractFile           string `yaml:"extract-file" mapstructure:"extract-file"`
}

func DefaultUIKeybinding() UIKeybindings {
	var result UIKeybindings
	defaults := key.DefaultBindings()

	// converts from key.Bindings to UIKeybindings
	getUIBindingValues(reflect.ValueOf(defaults), reflect.ValueOf(&result).Elem())

	return result
}

func getUIBindingValues(src, dst reflect.Value) {
	switch src.Kind() {
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			srcField := src.Field(i)
			srcType := src.Type().Field(i)

			if !srcField.CanInterface() {
				continue
			}

			dstField := dst.FieldByName(srcType.Name)
			if !dstField.IsValid() || !dstField.CanSet() {
				continue
			}

			if srcType.Type.Name() == "Config" {
				inputField := srcField.FieldByName("Input")
				if inputField.IsValid() && dstField.Kind() == reflect.String {
					dstField.SetString(inputField.String())
				}
				continue
			}
			getUIBindingValues(srcField, dstField)
		}
	}
}

func (c *UIKeybindings) PostLoad() error {
	cfg := key.Bindings{}

	// convert UIKeybindings to key.Bindings
	err := createKeyBindings(reflect.ValueOf(c).Elem(), reflect.ValueOf(&cfg).Elem())
	if err != nil {
		return err
	}

	c.Config = cfg
	return nil
}

func createKeyBindings(src, dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.Struct:
		for i := 0; i < dst.NumField(); i++ {
			dstField := dst.Field(i)
			dstType := dst.Type().Field(i)

			if !dstField.CanSet() {
				continue
			}

			srcField := src.FieldByName(dstType.Name)
			if !srcField.IsValid() {
				continue
			}

			if dstType.Type.Name() == "Config" {
				inputField := dstField.FieldByName("Input")
				if inputField.IsValid() && inputField.CanSet() && srcField.Kind() == reflect.String {
					inputField.SetString(srcField.String())

					// call the Setup method if it exists
					setupMethod := dstField.Addr().MethodByName("Setup")
					if setupMethod.IsValid() {
						result := setupMethod.Call([]reflect.Value{})
						if !result[0].IsNil() {
							return result[0].Interface().(error)
						}
					}
				}
				continue
			}
			err := createKeyBindings(srcField, dstField)
			if err != nil {
				return err
			}
		}
	}

	return nil
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
