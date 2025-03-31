package options

import (
	"github.com/anchore/clio"
	log "github.com/sirupsen/logrus"
	v1 "github.com/wagoodman/dive/runtime/ui/v1"
)

var _ interface {
	clio.PostLoader
	clio.FieldDescriber
} = (*UIDiff)(nil)

// UIDiff provides configuration for how differences are displayed
type UIDiff struct {
	Hide []string `yaml:"hide" mapstructure:"hide"`
}

func DefaultUIDiff() UIDiff {
	prefs := v1.DefaultPreferences()
	return UIDiff{
		Hide: prefs.FiletreeDiffHide,
	}
}

func (c *UIDiff) DescribeFields(descriptions clio.FieldDescriptionSet) {
	descriptions.Add(&c.Hide, "types of file differences to hide (added, removed, modified, unmodified)")
}

func (c *UIDiff) PostLoad() error {
	validHideValues := map[string]bool{"added": true, "removed": true, "modified": true, "unmodified": true}
	for _, value := range c.Hide {
		if _, ok := validHideValues[value]; !ok {
			log.Warnf("invalid diff hide value: %s (valid values: added, removed, modified, unmodified)", value)
		}
	}

	return nil
}
