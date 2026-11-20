package options

import (
	"strings"

	"github.com/anchore/clio"
	"github.com/fatih/color"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui/format"
	"github.com/wagoodman/dive/internal/log"
)

var _ interface {
	clio.PostLoader
	clio.FieldDescriber
} = (*UIColors)(nil)

// UIColors provides configuration for UI color customizations
type UIColors struct {
	Selected   string `yaml:"selected" mapstructure:"selected"`
	Added      string `yaml:"added" mapstructure:"added"`
	Removed    string `yaml:"removed" mapstructure:"removed"`
	Modified   string `yaml:"modified" mapstructure:"modified"`
	Unmodified string `yaml:"unmodified" mapstructure:"unmodified"`
	Layer      string `yaml:"layer" mapstructure:"layer"`
	Header     string `yaml:"header" mapstructure:"header"`
}

func DefaultUIColors() UIColors {
	return UIColors{}
}

func (c *UIColors) DescribeFields(descriptions clio.FieldDescriptionSet) {
	descriptions.Add(&c.Selected, "color for selected items in the UI (e.g. bright-white, reverse, bold)")
	descriptions.Add(&c.Added, "color for added files in the filetree (e.g. green)")
	descriptions.Add(&c.Removed, "color for removed files in the filetree (e.g. red)")
	descriptions.Add(&c.Modified, "color for modified files in the filetree (e.g. yellow)")
	descriptions.Add(&c.Unmodified, "color for unmodified files in the filetree")
	descriptions.Add(&c.Layer, "color for layer items in the layer pane (e.g. yellow)")
	descriptions.Add(&c.Header, "color for pane headers (e.g. bold)")
}

func (c *UIColors) PostLoad() error {
	validate := func(name, val string) {
		if strings.TrimSpace(val) == "" {
			return
		}
		if _, err := format.ParseColor(val); err != nil {
			log.Warnf("invalid color for %s: %s (%v)", name, val, err)
		}
	}
	validate("selected", c.Selected)
	validate("added", c.Added)
	validate("removed", c.Removed)
	validate("modified", c.Modified)
	validate("unmodified", c.Unmodified)
	validate("layer", c.Layer)
	validate("header", c.Header)
	return nil
}

func (c UIColors) ToTUIColors() tui.Colors {
	parse := func(val string) *color.Color {
		if strings.TrimSpace(val) == "" {
			return nil
		}
		col, err := format.ParseColor(val)
		if err != nil {
			return nil
		}
		return col
	}

	return tui.Colors{
		Selected:   parse(c.Selected),
		Added:      parse(c.Added),
		Removed:    parse(c.Removed),
		Modified:   parse(c.Modified),
		Unmodified: parse(c.Unmodified),
		Layer:      parse(c.Layer),
		Header:     parse(c.Header),
	}
}
