package config

type keybindingConfig struct {
	Quit                     string `mapstructure:"quit"`
	ToggleViews              string `mapstructure:"toggle-view"`
	FilterFiles              string `mapstructure:"filter-files"`
	CompareAll               string `mapstructure:"compare-all"`
	CompareLayer             string `mapstructure:"compare-layer"`
	ToggleCollapseDir        string `mapstructure:"toggle-collapse-dir"`
	ToggleCollapseAllDir     string `mapstructure:"toggle-collapse-all-dir"`
	ToggleFileTreeAttributes string `mapstructure:"toggle-filetree-attributes"`
	ToggleAddedFiles         string `mapstructure:"toggle-added-files"`
	ToggleRemovedFiles       string `mapstructure:"toggle-removed-files"`
	ToggleModifiedFiles      string `mapstructure:"toggle-modified-files"`
	ToggleUnmodifiedFiles    string `mapstructure:"toggle-unmodified-files"`
	ToggleWrapTree           string `mapstructure:"toggle-wrap-tree"`
	PageUp                   string `mapstructure:"page-up"`
	PageDown                 string `mapstructure:"page-down"`
}
