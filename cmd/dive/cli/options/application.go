package options

import v1 "github.com/wagoodman/dive/runtime/ui/v1"

type Application struct {
	Analysis Analysis `yaml:",inline" mapstructure:",squash"`
	CI       CI       `yaml:",inline" mapstructure:",squash"`
	Export   Export   `yaml:",inline" mapstructure:",squash"`
	UI       UI       `yaml:",inline" mapstructure:",squash"`
}

func DefaultApplication() Application {
	return Application{
		Analysis: DefaultAnalysis(),
		CI:       DefaultCI(),
		Export:   DefaultExport(),
		UI:       DefaultUI(),
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
