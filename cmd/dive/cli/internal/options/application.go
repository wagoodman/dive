package options

import (
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1"
)

type Application struct {
	Development Development `yaml:"dev" mapstructure:"dev"`
	Analysis    Analysis    `yaml:",inline" mapstructure:",squash"`
	CI          CI          `yaml:",inline" mapstructure:",squash"`
	Export      Export      `yaml:",inline" mapstructure:",squash"`
	UI          UI          `yaml:",inline" mapstructure:",squash"`
}

func DefaultApplication() Application {
	return Application{
		Development: DefaultDevelopment(),
		Analysis:    DefaultAnalysis(),
		CI:          DefaultCI(),
		Export:      DefaultExport(),
		UI:          DefaultUI(),
	}
}

func DefaultDevelopment() Development {
	return Development{
		UseStereoscope: false,
	}
}

func (c Application) V1Preferences() v1.Preferences {
	return v1.Preferences{
		KeyBindings:                c.UI.Keybinding.Config,
		ShowFiletreeAttributes:     c.UI.Filetree.ShowAttributes,
		ShowAggregatedLayerChanges: c.UI.Layer.ShowAggregatedChanges,
		CollapseFiletreeDirectory:  c.UI.Filetree.CollapseDir,
		FiletreePaneWidth:          c.UI.Filetree.PaneWidth,
		FiletreeDiffHide:           nil,
	}
}

type Development struct {
	UseStereoscope bool `yaml:"use-stereoscope" mapstructure:"use-stereoscope"`
}
