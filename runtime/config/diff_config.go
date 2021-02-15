package config

import "github.com/wagoodman/dive/dive/filetree"

type diffConfig struct {
	// note: this really relates to the fileTreeViewConfig, but is here for legacy reasons
	Hide      []string            `mapstructure:"hide"`
	DiffTypes []filetree.DiffType `yaml:"-"`
}

func (c *diffConfig) build() error {
	c.DiffTypes = nil
	for _, hideType := range c.Hide {
		switch hideType {
		case "added":
			c.DiffTypes = append(c.DiffTypes, filetree.Added)
		case "removed":
			c.DiffTypes = append(c.DiffTypes, filetree.Removed)
		case "modified":
			c.DiffTypes = append(c.DiffTypes, filetree.Modified)
		case "unmodified":
			c.DiffTypes = append(c.DiffTypes, filetree.Unmodified)
		}
	}
	return nil
}
