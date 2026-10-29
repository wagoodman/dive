package options

import (
	"fmt"
	"strings"

	"github.com/anchore/clio"

	"github.com/wagoodman/dive/dive/source"
	"github.com/wagoodman/dive/internal/log"
)

const defaultContainerEngine = "docker"

var _ interface {
	clio.PostLoader
	clio.FieldDescriber
} = (*Analysis)(nil)

// Analysis provides configuration for the image analysis behavior
type Analysis struct {
	// Image is the image argument exactly as the user gave it (scheme included), see source.Parse
	Image                     string   `yaml:"image" mapstructure:"-"`
	ContainerEngine           string   `yaml:"container-engine" mapstructure:"container-engine"`
	Platform                  string   `yaml:"platform" mapstructure:"platform"`
	IgnoreErrors              bool     `yaml:"ignore-errors" mapstructure:"ignore-errors"`
	AvailableContainerEngines []string `yaml:"-" mapstructure:"-"`
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

	return nil
}
