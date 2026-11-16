package options

import (
	"fmt"
	"os"
	"strings"

	"github.com/anchore/clio"
	"gopkg.in/yaml.v3"
)

var _ interface {
	clio.PostLoader
	clio.FieldDescriber
	clio.FlagAdder
} = (*CI)(nil)

const defaultCIConfigPath = ".dive-ci"

type CI struct {
	Enabled    bool    `yaml:"ci" mapstructure:"ci"`
	ConfigPath string  `yaml:"ci-config" mapstructure:"ci-config"`
	Rules      CIRules `yaml:"rules" mapstructure:"rules"`
}

func DefaultCI() CI {
	return CI{
		// the CI env var only sets the default, so --ci, DIVE_CI and the config file can still override it
		Enabled:    truthy(os.Getenv("CI")),
		ConfigPath: defaultCIConfigPath,
		Rules:      DefaultCIRules(),
	}
}

func (c *CI) DescribeFields(descriptions clio.FieldDescriptionSet) {
	descriptions.Add(&c.Enabled, "enable CI mode")
	descriptions.Add(&c.ConfigPath, "path to the CI config file")
}

func (c *CI) AddFlags(flags clio.FlagSet) {
	flags.BoolVarP(&c.Enabled, "ci", "", "skip the interactive TUI and validate against CI rules (same as env var CI=true)")
	flags.StringVarP(&c.ConfigPath, "ci-config", "", "yaml file with the CI validation rules, used with --ci or CI=true in the environment")
}

func (c *CI) PostLoad() error {
	if c.ConfigPath != "" {
		if fileExists(c.ConfigPath) {
			// if a config file is provided, load it and override any values provided in the application config.
			// If we're hitting this case we should pretend that only the config file was provided and applied
			// on top of the default config values.
			yamlFile, err := os.ReadFile(c.ConfigPath)
			if err != nil {
				return fmt.Errorf("failed to read CI config file %s: %w", c.ConfigPath, err)
			}
			def := DefaultCIRules()
			r := legacyRuleFile{
				LowestEfficiencyThresholdString: def.LowestEfficiencyThresholdString,
				HighestWastedBytesString:        def.HighestWastedBytesString,
				HighestUserWastedPercentString:  def.HighestUserWastedPercentString,
			}
			wrapper := struct {
				Rules *legacyRuleFile `yaml:"rules"`
			}{
				Rules: &r,
			}
			if err := yaml.Unmarshal(yamlFile, &wrapper); err != nil {
				return fmt.Errorf("failed to unmarshal CI config file %s: %w", c.ConfigPath, err)
			}
			// TODO: should this be a deprecated use warning in the future?
			c.Rules = CIRules{
				LowestEfficiencyThresholdString: r.LowestEfficiencyThresholdString,
				HighestWastedBytesString:        r.HighestWastedBytesString,
				HighestUserWastedPercentString:  r.HighestUserWastedPercentString,
			}
		}
	}

	return nil
}

type legacyRuleFile struct {
	LowestEfficiencyThresholdString string `yaml:"lowestEfficiency"`
	HighestWastedBytesString        string `yaml:"highestWastedBytes"`
	HighestUserWastedPercentString  string `yaml:"highestUserWastedPercent"`
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// truthy matches CI env values case-insensitively, e.g. AppVeyor sets CI=True
func truthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}
