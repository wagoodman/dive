package key

import (
	"fmt"
	"github.com/awesome-gocui/keybinding"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Input string
	Keys  []keybinding.Key
}

func (c *Config) Setup() error {
	if len(c.Input) == 0 {
		return nil
	}

	parsed, err := keybinding.ParseAll(c.Input)
	if err != nil {
		return fmt.Errorf("failed to parse key %q: %w", c.Input, err)
	}
	c.Keys = parsed
	return nil
}

func (c *Config) UnmarshalText(text []byte) error {
	c.Input = string(text)

	return nil
}

func (c Config) MarshalText() ([]byte, error) {
	return []byte(c.Input), nil
}

func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	var strValue string
	if err := value.Decode(&strValue); err != nil {
		return err
	}

	return c.UnmarshalText([]byte(strValue))
}

func (c Config) MarshalYAML() (interface{}, error) {
	return c.Input, nil
}

type Bindings struct {
	Global     GlobalBindings     `yaml:",inline" mapstructure:",squash"`
	Navigation NavigationBindings `yaml:",inline" mapstructure:",squash"`
	Layer      LayerBindings      `yaml:",inline" mapstructure:",squash"`
	Filetree   FiletreeBindings   `yaml:",inline" mapstructure:",squash"`
}

type GlobalBindings struct {
	Quit             Config `yaml:"quit" mapstructure:"quit"`
	ToggleView       Config `yaml:"toggle-view" mapstructure:"toggle-view"`
	FilterFiles      Config `yaml:"filter-files" mapstructure:"filter-files"`
	CloseFilterFiles Config `yaml:"close-filter-files" mapstructure:"close-filter-files"`
}

type NavigationBindings struct {
	Up       Config `yaml:"up" mapstructure:"up"`
	Down     Config `yaml:"down" mapstructure:"down"`
	Left     Config `yaml:"left" mapstructure:"left"`
	Right    Config `yaml:"right" mapstructure:"right"`
	PageUp   Config `yaml:"page-up" mapstructure:"page-up"`
	PageDown Config `yaml:"page-down" mapstructure:"page-down"`
}

type LayerBindings struct {
	CompareAll   Config `yaml:"compare-all" mapstructure:"compare-all"`
	CompareLayer Config `yaml:"compare-layer" mapstructure:"compare-layer"`
}

type FiletreeBindings struct {
	ToggleCollapseDir     Config `yaml:"toggle-collapse-dir" mapstructure:"toggle-collapse-dir"`
	ToggleCollapseAllDir  Config `yaml:"toggle-collapse-all-dir" mapstructure:"toggle-collapse-all-dir"`
	ToggleAddedFiles      Config `yaml:"toggle-added-files" mapstructure:"toggle-added-files"`
	ToggleRemovedFiles    Config `yaml:"toggle-removed-files" mapstructure:"toggle-removed-files"`
	ToggleModifiedFiles   Config `yaml:"toggle-modified-files" mapstructure:"toggle-modified-files"`
	ToggleUnmodifiedFiles Config `yaml:"toggle-unmodified-files" mapstructure:"toggle-unmodified-files"`
	ToggleTreeAttributes  Config `yaml:"toggle-filetree-attributes" mapstructure:"toggle-filetree-attributes"`
	ToggleSortOrder       Config `yaml:"toggle-sort-order" mapstructure:"toggle-sort-order"`
	ToggleWrapTree        Config `yaml:"toggle-wrap-tree" mapstructure:"toggle-wrap-tree"`
	ExtractFile           Config `yaml:"extract-file" mapstructure:"extract-file"`
}
