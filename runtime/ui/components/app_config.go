package components

import (
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/dive/filetree"
)

type AppConfig struct {}

func (a *AppConfig) GetDefaultHide() (result []filetree.DiffType) {
	slice := viper.GetStringSlice("diff.hide")
	for _, hideType := range slice {
		switch hideType{
		case "added":
			result = append(result, filetree.Added)
		case "removed":
			result = append(result, filetree.Removed)
		case "modified":
			result = append(result, filetree.Modified)
		case "unmodified":
			result = append(result, filetree.Unmodified)
		}
	}

	return result
}

func (a *AppConfig) GetAggregateLayerSetting() bool {
	return viper.GetBool("layer.show-aggregated-changes")
}

func (a *AppConfig) GetCollapseDir() bool {
	return viper.GetBool("filetree.collapse-dir")
}

func (a *AppConfig) GetPaneWidth() (int,int) {
	fwp := viper.GetFloat64("filetree.pane-width")
	lwp := 1 - fwp
	return int(fwp*100), int(lwp*100)
}

func (a *AppConfig) GetShowAttributes() bool {
	return viper.GetBool("filetree.show-attributes")
}
