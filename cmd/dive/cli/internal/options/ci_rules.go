package options

import (
	"os"
	"strings"

	"github.com/anchore/clio"
	"github.com/spf13/pflag"

	"github.com/wagoodman/dive/cmd/dive/cli/internal/command/ci"
	"github.com/wagoodman/dive/internal/log"
)

type CIRules struct {
	LowestEfficiencyThresholdString       string `yaml:"lowest-efficiency" mapstructure:"lowest-efficiency"`
	LegacyLowestEfficiencyThresholdString string `yaml:"-" mapstructure:"lowestEfficiency"`

	HighestWastedBytesString       string `yaml:"highest-wasted-bytes" mapstructure:"highest-wasted-bytes"`
	LegacyHighestWastedBytesString string `yaml:"-" mapstructure:"highestWastedBytes"`

	HighestUserWastedPercentString       string `yaml:"highest-user-wasted-percent" mapstructure:"highest-user-wasted-percent"`
	LegacyHighestUserWastedPercentString string `yaml:"-" mapstructure:"highestUserWastedPercent"`

	List []ci.Rule `yaml:"-" mapstructure:"-"`

	// flags is where the rule flags were registered, to tell which ones were given on the command line
	flags *pflag.FlagSet
}

func DefaultCIRules() CIRules {
	return CIRules{
		LowestEfficiencyThresholdString: "0.9",
		HighestWastedBytesString:        "disabled",
		HighestUserWastedPercentString:  "0.1",
	}
}

func (c *CIRules) DescribeFields(descriptions clio.FieldDescriptionSet) {
	descriptions.Add(&c.LowestEfficiencyThresholdString, "lowest allowable image efficiency (as a ratio between 0-1), otherwise CI validation will fail.")
	descriptions.Add(&c.HighestWastedBytesString, "highest allowable bytes wasted, otherwise CI validation will fail.")
	descriptions.Add(&c.HighestUserWastedPercentString, "highest allowable percentage of bytes wasted (as a ratio between 0-1), otherwise CI validation will fail.")
}

func (c *CIRules) AddFlags(flags clio.FlagSet) {
	flags.StringVarP(&c.LowestEfficiencyThresholdString, "lowestEfficiency", "", "(only valid with --ci given) lowest allowable image efficiency (as a ratio between 0-1), otherwise CI validation will fail.")
	flags.StringVarP(&c.HighestWastedBytesString, "highestWastedBytes", "", "(only valid with --ci given) highest allowable bytes wasted, otherwise CI validation will fail.")
	flags.StringVarP(&c.HighestUserWastedPercentString, "highestUserWastedPercent", "", "(only valid with --ci given) highest allowable percentage of bytes wasted (as a ratio between 0-1), otherwise CI validation will fail.")

	if p, ok := flags.(interface{ PFlagSet() *pflag.FlagSet }); ok {
		c.flags = p.PFlagSet()
	}
}

// ruleKey names one rule in each of its spellings, camel is also the flag name
type ruleKey struct {
	snake, camel  string
	value, legacy *string
}

func (c *CIRules) keys() []ruleKey {
	return []ruleKey{
		{snake: "lowest-efficiency", camel: "lowestEfficiency", value: &c.LowestEfficiencyThresholdString, legacy: &c.LegacyLowestEfficiencyThresholdString},
		{snake: "highest-wasted-bytes", camel: "highestWastedBytes", value: &c.HighestWastedBytesString, legacy: &c.LegacyHighestWastedBytesString},
		{snake: "highest-user-wasted-percent", camel: "highestUserWastedPercent", value: &c.HighestUserWastedPercentString, legacy: &c.LegacyHighestUserWastedPercentString},
	}
}

// explicit reports whether the rule was set by its flag or env var, which outrank any config file
func (c *CIRules) explicit(r ruleKey) bool {
	if c.flags != nil && c.flags.Changed(r.camel) {
		return true
	}
	_, ok := os.LookupEnv("DIVE_RULES_" + strings.ToUpper(strings.ReplaceAll(r.snake, "-", "_")))
	return ok
}

func (c CIRules) hasLegacyOptionsInUse() bool {
	return c.LegacyLowestEfficiencyThresholdString != "" || c.LegacyHighestWastedBytesString != "" || c.LegacyHighestUserWastedPercentString != ""
}

func (c *CIRules) PostLoad() error {
	// protect against repeated calls
	c.List = nil

	if c.hasLegacyOptionsInUse() {
		log.Warnf("please specify ci rules in snake-case (the legacy camelCase format is deprecated)")
	}

	// the legacy keys only come from config files, so a flag or env var still wins over them
	for _, r := range c.keys() {
		if *r.legacy != "" && !c.explicit(r) {
			*r.value = *r.legacy
		}
	}

	rules, err := ci.Rules(c.LowestEfficiencyThresholdString, c.HighestWastedBytesString, c.HighestUserWastedPercentString)
	if err != nil {
		return err
	}
	c.List = append(c.List, rules...)

	return nil
}
