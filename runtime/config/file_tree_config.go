package config

type fileTreeViewConfig struct {
	CollapseDir    bool    `mapstructure:"collapse-dir"`
	PaneWidthRatio float64 `mapstructure:"pane-width"`
	ShowAttributes bool    `mapstructure:"show-attributes"`
}
