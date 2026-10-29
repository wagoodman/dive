package options

import (
	"fmt"
	"github.com/anchore/clio"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/dive/source"
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
	Image string `yaml:"image" mapstructure:"-"`
	// Reference is the image argument exactly as the user gave it (scheme included), for source.Load
	Reference       string `yaml:"-" mapstructure:"-"`
	ContainerEngine string `yaml:"container-engine" mapstructure:"container-engine"`
	Platform        string `yaml:"platform" mapstructure:"platform"`
	// Source is only used by the old resolver path, it knows docker, podman and docker-archive
	Source                    dive.ImageSource `yaml:"-" mapstructure:"-"`
	IgnoreErrors              bool             `yaml:"ignore-errors" mapstructure:"ignore-errors"`
	AvailableContainerEngines []string         `yaml:"-" mapstructure:"-"`
}

func DefaultAnalysis() Analysis {
	return Analysis{
		ContainerEngine:           defaultContainerEngine,
		IgnoreErrors:              false,
		AvailableContainerEngines: source.Sources,
	}
}

func (c *Analysis) DescribeFields(descriptions clio.FieldDescriptionSet) {
	descriptions.Add(&c.ContainerEngine, fmt.Sprintf("where to load the image from when the image reference has no scheme (supported options: %s)", strings.Join(c.AvailableContainerEngines, ", ")))
	descriptions.Add(&c.Platform, "platform to select from a multi-platform image, e.g. 'linux/arm64' (default: the image source's default)")
	descriptions.Add(&c.IgnoreErrors, "continue with analysis even if there are errors parsing the image archive")
}

func (c *Analysis) AddFlags(flags clio.FlagSet) {
	flags.StringVarP(&c.ContainerEngine, "source", "",
		fmt.Sprintf("The container engine to fetch the image from. Allowed values: %s", strings.Join(c.AvailableContainerEngines, ", ")))

	flags.StringVarP(&c.Platform, "platform", "", "platform to select from a multi-platform image (e.g. 'linux/arm64')")

	flags.BoolVarP(&c.IgnoreErrors, "ignore-errors", "i", "ignore image parsing errors and run the analysis anyway")
}

func (c *Analysis) PostLoad() error {
	if _, ok := source.Normalize(c.ContainerEngine); !ok {
		log.Warnf("invalid container engine: %s (valid options: %s), using default %q", c.ContainerEngine, strings.Join(c.AvailableContainerEngines, ", "), defaultContainerEngine)
		c.ContainerEngine = defaultContainerEngine
	}

	c.Reference = c.Image

	// the rest maps onto the old resolver path, which only knows docker, podman and docker-archive
	if c.Image != "" {
		sourceType, imageStr := dive.DeriveImageSource(c.Image)

		if sourceType == dive.SourceUnknown {
			// new sources stay unknown here, GetImageResolver reports them
			sourceType = dive.ParseImageSource(c.ContainerEngine)

			// use exactly what the user provided
			imageStr = c.Image
		}

		c.Image = imageStr
		c.Source = sourceType
	} else {
		c.Source = dive.ParseImageSource(c.ContainerEngine)
	}

	return nil
}
