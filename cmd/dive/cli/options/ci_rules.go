package options

import (
	"github.com/anchore/clio"
	"github.com/wagoodman/dive/runtime/ci"
)

type CIRules struct {
	// TODO: allow for snake or camel case

	// user values
	LowestEfficiencyThresholdString string `yaml:"lowest-efficiency-threshold" mapstructure:"lowest-efficiency-threshold"`
	HighestWastedBytesString        string `yaml:"highest-wasted-bytes" mapstructure:"highest-wasted-bytes"`
	HighestUserWastedPercentString  string `yaml:"highest-user-wasted-percent" mapstructure:"highest-user-wasted-percent"`

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

func (c *CIRules) PostLoad() error {
	// protect against repeated calls
	c.List = nil

	rules, err := ci.Rules(c.LowestEfficiencyThresholdString, c.HighestWastedBytesString, c.HighestUserWastedPercentString)
	if err != nil {
		return err
	}
	c.List = append(c.List, rules...)

	return nil
}
