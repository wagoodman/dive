package v1

import (
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/key"
	"github.com/wagoodman/dive/dive"
)

type Config struct {
	// required input
	Analysis    *dive.Analysis
	Preferences Preferences
}

type Preferences struct {
	KeyBindings                key.Bindings
	IgnoreErrors               bool
	ShowFiletreeAttributes     bool
	ShowAggregatedLayerChanges bool
	CollapseFiletreeDirectory  bool
	FiletreePaneWidth          float64
	FiletreeDiffHide           []string
}

func DefaultPreferences() Preferences {
	return Preferences{
		KeyBindings:                key.DefaultBindings(),
		ShowFiletreeAttributes:     true,
		ShowAggregatedLayerChanges: true,
		CollapseFiletreeDirectory:  false, // don't start with collapsed directories
		FiletreePaneWidth:          0.5,
		FiletreeDiffHide:           []string{}, // empty slice means show all
	}
}
