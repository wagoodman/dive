package options

import (
	"fmt"
	"github.com/anchore/clio"
	"gopkg.in/yaml.v3"
	"os"
)

var _ interface {
	clio.PostLoader
	clio.FieldDescriber
	clio.FlagAdder
} = (*CI)(nil)

type CI struct {
	Enabled    bool    `yaml:"ci" mapstructure:"ci"`
	ConfigPath string  `yaml:"ci-config" mapstructure:"ci-config"`
	Rules      CIRules `yaml:"rules" mapstructure:"rules"`
}

func DefaultCI() CI {
	return CI{
		Enabled:    false,
		ConfigPath: ".dive-ci",
		Rules:      DefaultCIRules(),
	}
}

func (c *CI) DescribeFields(descriptions clio.FieldDescriptionSet) {
	descriptions.Add(&c.Enabled, "enable CI mode")
	descriptions.Add(&c.ConfigPath, "path to the CI config file")
}

func (c *CI) AddFlags(flags clio.FlagSet) {
	flags.BoolVarP(&c.Enabled, "ci", "", "Skip the interactive TUI and validate against CI rules (same as env var CI=true)")
	flags.StringVarP(&c.ConfigPath, "ci-config", "", "If CI=true in the environment, use the given yaml to drive validation rules.")
}

func (c *CI) PostLoad() error {
	enabledFromEnv := truthy(os.Getenv("CI"))
	if !c.Enabled && enabledFromEnv {
		c.Enabled = true
	}

	if c.ConfigPath != "" && fileExists(c.ConfigPath) {
		// if a config file is provided, load it and override any values provided in the application config.
		// If we're hitting this case we should pretend that only the config file was provided and applied
		// on top of the default config values.
		yamlFile, err := os.ReadFile(c.ConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read CI config file %s: %w", c.ConfigPath, err)
		}
		c.Rules = DefaultCIRules()
		if err := yaml.Unmarshal(yamlFile, &c.Rules); err != nil {
			return fmt.Errorf("failed to unmarshal CI config file %s: %w", c.ConfigPath, err)
		}
	}
	return nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func truthy(value string) bool {
	switch value {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return false
	}
}
