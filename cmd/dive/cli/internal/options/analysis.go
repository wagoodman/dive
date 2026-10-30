package options

import (
	"fmt"
	"strings"

	"github.com/anchore/clio"
	"github.com/anchore/stereoscope/pkg/image"

	"github.com/wagoodman/dive/internal/log"
	"github.com/wagoodman/dive/internal/source"
)

const defaultSource = "docker"

var _ interface {
	clio.PostLoader
	clio.FieldDescriber
} = (*Analysis)(nil)

// Analysis provides configuration for the image analysis behavior
type Analysis struct {
	// Image is the image argument exactly as the user gave it (scheme included), see source.Parse
	Image string `yaml:"image" mapstructure:"-"`
	// Source is the image source for references without a scheme. The config key predates the sources that aren't
	// container engines (archives, registries) and stays for compatibility
	Source   string `yaml:"container-engine" mapstructure:"container-engine"`
	Platform string `yaml:"platform" mapstructure:"platform"`
	// IgnoreErrors is deprecated and unused
	IgnoreErrors     bool     `yaml:"ignore-errors" mapstructure:"ignore-errors"`
	AvailableSources []string `yaml:"-" mapstructure:"-"`
}

func DefaultAnalysis() Analysis {
	return Analysis{
		Source:           defaultSource,
		IgnoreErrors:     false,
		AvailableSources: source.Sources,
	}
}

func (c *Analysis) DescribeFields(descriptions clio.FieldDescriptionSet) {
	descriptions.Add(&c.Source, fmt.Sprintf("where to load the image from when the image reference has no scheme (supported options: %s)", strings.Join(c.AvailableSources, ", ")))
	descriptions.Add(&c.Platform, "platform to select from a multi-platform image, e.g. 'linux/arm64' (default: the image source's default)")
	descriptions.Add(&c.IgnoreErrors, "deprecated, has no effect: image path errors no longer stop the analysis or the TUI")
}

func (c *Analysis) AddFlags(flags clio.FlagSet) {
	flags.StringVarP(&c.Source, "source", "",
		fmt.Sprintf("where to load the image from when the reference has no scheme (an engine, archive, dir or registry). Allowed values: %s", strings.Join(c.AvailableSources, ", ")))

	flags.StringVarP(&c.Platform, "platform", "", "platform to select from a multi-platform image (e.g. 'linux/arm64')")

	// ponytail: kept as a no-op so existing scripts and configs still parse
	flags.BoolVarP(&c.IgnoreErrors, "ignore-errors", "i", "(deprecated, has no effect) ignore image parsing errors and run the analysis anyway")
}

func (c *Analysis) PostLoad() error {
	if _, ok := source.Normalize(c.Source); !ok {
		log.Warnf("invalid image source: %s (valid options: %s), using default %q", c.Source, strings.Join(c.AvailableSources, ", "), defaultSource)
		c.Source = defaultSource
	}
	if c.Platform != "" {
		if _, err := image.NewPlatform(c.Platform); err != nil {
			return fmt.Errorf("invalid platform %q: %w", c.Platform, err)
		}
	}
	if c.IgnoreErrors {
		log.Warn("ignore-errors is deprecated and has no effect")
	}

	return nil
}
