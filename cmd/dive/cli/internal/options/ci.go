package options

import (
	"fmt"
	"os"
	"strings"

	"github.com/anchore/clio"
	"gopkg.in/yaml.v3"

	"github.com/wagoodman/dive/internal/log"
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
	// the rules file only matters in CI mode, so a broken one can't break the TUI or --json
	if !c.Enabled {
		return nil
	}

	// only the default file is optional, a path the user gave has to exist
	if c.ConfigPath != "" && c.ConfigPath != defaultCIConfigPath && !fileExists(c.ConfigPath) {
		return fmt.Errorf("CI config file does not exist: %s", c.ConfigPath)
	}

	if c.ConfigPath != "" {
		if fileExists(c.ConfigPath) {
			// a file with rules is the whole rule set: it replaces the app config rules and the defaults, and rules
			// it leaves out are skipped. Rule flags and env vars still win over it, as they do over any config file
			yamlFile, err := os.ReadFile(c.ConfigPath)
			if err != nil {
				return fmt.Errorf("failed to read CI config file %s: %w", c.ConfigPath, err)
			}
			var file struct {
				Rules map[string]string `yaml:"rules"`
			}
			if err := yaml.Unmarshal(yamlFile, &file); err != nil {
				return fmt.Errorf("failed to unmarshal CI config file %s: %w", c.ConfigPath, err)
			}
			if len(file.Rules) == 0 {
				return nil
			}
			for _, r := range c.Rules.keys() {
				v := "disabled"
				// the file has always used camelCase keys, snake-case (like the app config) wins when both are set
				for _, key := range []string{r.camel, r.snake} {
					if fv, ok := file.Rules[key]; ok {
						v = fv
					}
					delete(file.Rules, key)
				}
				if c.Rules.explicit(r) {
					continue
				}
				*r.value = v
				*r.legacy = ""
			}
			for key := range file.Rules {
				log.Warnf("unknown rule %q in CI config file %s", key, c.ConfigPath)
			}
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

// truthy matches CI env values case-insensitively, e.g. AppVeyor sets CI=True
func truthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}
