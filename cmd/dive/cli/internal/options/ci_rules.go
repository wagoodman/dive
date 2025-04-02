package options

import (
	"github.com/anchore/clio"
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

	if c.LegacyLowestEfficiencyThresholdString != "" {
		c.LowestEfficiencyThresholdString = c.LegacyLowestEfficiencyThresholdString
	}

	if c.LegacyHighestWastedBytesString != "" {
		c.HighestWastedBytesString = c.LegacyHighestWastedBytesString
	}

	if c.LegacyHighestUserWastedPercentString != "" {
		c.HighestUserWastedPercentString = c.LegacyHighestUserWastedPercentString
	}

	rules, err := ci.Rules(c.LowestEfficiencyThresholdString, c.HighestWastedBytesString, c.HighestUserWastedPercentString)
	if err != nil {
		return err
	}
	c.List = append(c.List, rules...)

	return nil
}
