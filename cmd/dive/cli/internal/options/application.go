package options

import (
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
)

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

func (c Application) TUIPreferences() tui.Preferences {
	return tui.Preferences{
		KeyBindings:                c.UI.Keybinding.Config,
		ShowFiletreeAttributes:     c.UI.Filetree.ShowAttributes,
		ShowAggregatedLayerChanges: c.UI.Layer.ShowAggregatedChanges,
		CollapseFiletreeDirectory:  c.UI.Filetree.CollapseDir,
		FiletreePaneWidth:          c.UI.Filetree.PaneWidth,
		FiletreeDiffHide:           nil,
		Colors:                     c.UI.Colors.ToTUIColors(),
	}
}
