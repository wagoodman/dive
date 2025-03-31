package options

import (
	"github.com/anchore/clio"
	log "github.com/sirupsen/logrus"
	v1 "github.com/wagoodman/dive/runtime/ui/v1"
)

var _ interface {
	clio.PostLoader
	clio.FieldDescriber
} = (*UIFiletree)(nil)

// UIFiletree provides configuration for the file tree display
type UIFiletree struct {
	CollapseDir    bool    `yaml:"collapse-dir" mapstructure:"collapse-dir"`
	PaneWidth      float64 `yaml:"pane-width" mapstructure:"pane-width"`
	ShowAttributes bool    `yaml:"show-attributes" mapstructure:"show-attributes"`
}

func DefaultUIFiletree() UIFiletree {
	prefs := v1.DefaultPreferences()
	return UIFiletree{
		CollapseDir:    prefs.CollapseFiletreeDirectory,
		PaneWidth:      prefs.FiletreePaneWidth,
		ShowAttributes: prefs.ShowFiletreeAttributes,
	}
}

func (c *UIFiletree) DescribeFields(descriptions clio.FieldDescriptionSet) {
	descriptions.Add(&c.CollapseDir, "collapse directories by default in the filetree")
	descriptions.Add(&c.PaneWidth, "percentage of screen width for the filetree pane (must be >0 and <1)")
	descriptions.Add(&c.ShowAttributes, "show file attributes in the filetree view")
}

func (c *UIFiletree) PostLoad() error {
	// Validate pane width is between 0 and 1
	if c.PaneWidth <= 0 || c.PaneWidth >= 1 {
		log.Warnf("filetree pane-width must be >0 and <1, got %v, resetting to default 0.5", c.PaneWidth)
		c.PaneWidth = 0.5
	}
	return nil
}
