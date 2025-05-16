package options

import (
	"fmt"
	"github.com/anchore/clio"
	"github.com/scylladb/go-set/strset"
	diveV1 "github.com/wagoodman/dive/dive/v1"
	"github.com/wagoodman/dive/internal/log"
	"strings"
)

const defaultContainerEngine = "docker"

var _ interface {
	clio.PostLoader
	clio.FieldDescriber
} = (*Analysis)(nil)

// Analysis provides configuration for the image analysis behavior
type Analysis struct {
	Image                     string             `yaml:"image" mapstructure:"-"`
	ContainerEngine           string             `yaml:"container-engine" mapstructure:"container-engine"`
	Source                    diveV1.ImageSource `yaml:"-" mapstructure:"-"`
	IgnoreErrors              bool               `yaml:"ignore-errors" mapstructure:"ignore-errors"`
	AvailableContainerEngines []string           `yaml:"-" mapstructure:"-"`
}

func DefaultAnalysis() Analysis {
	return Analysis{
		ContainerEngine:           defaultContainerEngine,
		IgnoreErrors:              false,
		AvailableContainerEngines: diveV1.ImageSources,
	}
}

func (c *Analysis) DescribeFields(descriptions clio.FieldDescriptionSet) {
	descriptions.Add(&c.ContainerEngine, "container engine to use for image analysis (supported options: 'docker' and 'podman')")
	descriptions.Add(&c.IgnoreErrors, "continue with analysis even if there are errors parsing the image archive")
}

func (c *Analysis) AddFlags(flags clio.FlagSet) {
	flags.StringVarP(&c.ContainerEngine, "source", "",
		fmt.Sprintf("The container engine to fetch the image from. Allowed values: %s", strings.Join(c.AvailableContainerEngines, ", ")))

	flags.BoolVarP(&c.IgnoreErrors, "ignore-errors", "i", "ignore image parsing errors and run the analysis anyway")
}

func (c *Analysis) PostLoad() error {
	validEngines := strset.New(c.AvailableContainerEngines...)
	if !validEngines.Has(c.ContainerEngine) {
		log.Warnf("invalid container engine: %s (valid options: %s), using default %q", c.ContainerEngine, strings.Join(c.AvailableContainerEngines, ", "), defaultContainerEngine)
		c.ContainerEngine = "docker"
	}

	if c.Image != "" {
		sourceType, imageStr := diveV1.DeriveImageSource(c.Image)

		if sourceType == diveV1.SourceUnknown {
			sourceType = diveV1.ParseImageSource(c.ContainerEngine)
			if sourceType == diveV1.SourceUnknown {
				return fmt.Errorf("unable to determine image source from %q: %v\n", c.Image, c.ContainerEngine)
			}

			// use exactly what the user provided
			imageStr = c.Image
		}

		c.Image = imageStr
		c.Source = sourceType
	} else {
		c.Source = diveV1.ParseImageSource(c.ContainerEngine)
	}

	return nil
}
