package tui

import (
	"github.com/fatih/color"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui/filetree"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui/format"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui/key"
	"github.com/wagoodman/dive/dive"
)

type Config struct {
	// required input
	Analysis    *dive.Analysis
	Preferences Preferences
}

type Colors struct {
	Selected   *color.Color
	Added      *color.Color
	Removed    *color.Color
	Modified   *color.Color
	Unmodified *color.Color
	Layer      *color.Color
	Header     *color.Color
}

func (c Colors) Apply() {
	if c.Selected != nil {
		format.SetSelected(c.Selected)
	}
	if c.Layer != nil {
		format.SetLayer(c.Layer)
	}
	if c.Header != nil {
		format.SetHeader(c.Header)
	}
	if c.Added != nil {
		filetree.SetDiffTypeColor(filetree.Added, c.Added)
	}
	if c.Removed != nil {
		filetree.SetDiffTypeColor(filetree.Removed, c.Removed)
	}
	if c.Modified != nil {
		filetree.SetDiffTypeColor(filetree.Modified, c.Modified)
	}
	if c.Unmodified != nil {
		filetree.SetDiffTypeColor(filetree.Unmodified, c.Unmodified)
	}
}

func ResetDefaultColors() {
	format.ResetDefaults()
	filetree.ResetDiffTypeColors()
}

type Preferences struct {
	KeyBindings                key.Bindings
	ShowFiletreeAttributes     bool
	ShowAggregatedLayerChanges bool
	CollapseFiletreeDirectory  bool
	FiletreePaneWidth          float64
	FiletreeDiffHide           []string
	Colors                     Colors
}

func DefaultPreferences() Preferences {
	return Preferences{
		KeyBindings:                key.DefaultBindings(),
		ShowFiletreeAttributes:     true,
		ShowAggregatedLayerChanges: true,
		CollapseFiletreeDirectory:  false, // don't start with collapsed directories
		FiletreePaneWidth:          0.5,
		FiletreeDiffHide:           []string{}, // empty slice means show all
		Colors:                     Colors{},
	}
}
