package options

import (
	"fmt"
	"github.com/anchore/clio"
	"github.com/dustin/go-humanize"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ci"
	"strconv"
	"strings"
)

const (
	ciKeyLowestEfficiencyThreshold = "lowestEfficiency"
	ciKeyHighestWastedBytes        = "highestWastedBytes"
	ciKeyHighestUserWastedPercent  = "highestUserWastedPercent"
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

	ruleProcessors := []func() error{
		c.processEfficiencyRule,
		c.processWastedBytesRule,
		c.processUserWastedPercentRule,
	}

	for _, processor := range ruleProcessors {
		if err := processor(); err != nil {
			return err
		}
	}

	return nil
}

func (c *CIRules) processEfficiencyRule() error {
	if isRuleDisabled(c.LowestEfficiencyThresholdString) {
		c.List = append(c.List, disabledRule(ciKeyLowestEfficiencyThreshold))
		return nil
	}

	threshold, err := strconv.ParseFloat(c.LowestEfficiencyThresholdString, 64)
	if err != nil {
		return fmt.Errorf("invalid %s config value, given %q: %v",
			ciKeyLowestEfficiencyThreshold, c.LowestEfficiencyThresholdString, err)
	}

	if threshold < 0 || threshold > 1 {
		return fmt.Errorf("%s config value is outside allowed range (0-1), given '%f'",
			ciKeyLowestEfficiencyThreshold, threshold)
	}

	c.List = append(c.List,
		newGenericRule(
			ciKeyLowestEfficiencyThreshold,
			c.LowestEfficiencyThresholdString,
			func(analysis *image.AnalysisResult) (ci.RuleStatus, string) {
				if threshold > analysis.Efficiency {
					return ci.RuleFailed, fmt.Sprintf(
						"image efficiency is too low (efficiency=%v < threshold=%v)",
						analysis.Efficiency, threshold)
				}
				return ci.RulePassed, ""
			},
		),
	)

	return nil
}

func (c *CIRules) processWastedBytesRule() error {
	if isRuleDisabled(c.HighestWastedBytesString) {
		c.List = append(c.List, disabledRule(ciKeyHighestWastedBytes))
		return nil
	}

	threshold, err := humanize.ParseBytes(c.HighestWastedBytesString)
	if err != nil {
		return fmt.Errorf("invalid highestWastedBytes config value, given %q: %v",
			c.HighestWastedBytesString, err)
	}

	c.List = append(c.List,
		newGenericRule(
			ciKeyHighestWastedBytes,
			c.HighestWastedBytesString,
			func(analysis *image.AnalysisResult) (ci.RuleStatus, string) {
				if analysis.WastedBytes > threshold {
					return ci.RuleFailed, fmt.Sprintf(
						"too many bytes wasted (wasted-bytes=%v > threshold=%v)",
						analysis.WastedBytes, threshold)
				}
				return ci.RulePassed, ""
			},
		),
	)

	return nil
}

func (c *CIRules) processUserWastedPercentRule() error {
	if isRuleDisabled(c.HighestUserWastedPercentString) {
		c.List = append(c.List, disabledRule(ciKeyHighestUserWastedPercent))
		return nil
	}

	threshold, err := strconv.ParseFloat(c.HighestUserWastedPercentString, 64)
	if err != nil {
		return fmt.Errorf("invalid highestUserWastedPercent config value, given %q: %v",
			c.HighestUserWastedPercentString, err)
	}

	if threshold < 0 || threshold > 1 {
		return fmt.Errorf("highestUserWastedPercent config value is outside allowed range (0-1), given '%f'",
			threshold)
	}

	c.List = append(c.List,
		newGenericRule(
			ciKeyHighestUserWastedPercent,
			c.HighestUserWastedPercentString,
			func(analysis *image.AnalysisResult) (ci.RuleStatus, string) {
				if analysis.WastedUserPercent > threshold {
					return ci.RuleFailed, fmt.Sprintf(
						"too many bytes wasted, relative to the user bytes added (%%-user-wasted-bytes=%v > threshold=%v)",
						analysis.WastedUserPercent, threshold)
				}
				return ci.RulePassed, ""
			},
		),
	)

	return nil
}

func isRuleDisabled(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "" || value == "disabled" || value == "off" || value == "false"
}

type genericRule struct {
	key         string
	configValue string
	evaluator   func(*image.AnalysisResult) (ci.RuleStatus, string)
}

func newGenericRule(key string, configValue string, evaluator func(*image.AnalysisResult) (ci.RuleStatus, string)) *genericRule {
	return &genericRule{
		key:         key,
		configValue: configValue,
		evaluator:   evaluator,
	}
}

func (rule *genericRule) Key() string {
	return rule.key
}

func (rule *genericRule) Configuration() string {
	return rule.configValue
}

func (rule *genericRule) Evaluate(result *image.AnalysisResult) (ci.RuleStatus, string) {
	return rule.evaluator(result)
}

func disabledRule(key string) *genericRule {
	return &genericRule{
		key:         key,
		configValue: "disabled",
		evaluator: func(_ *image.AnalysisResult) (ci.RuleStatus, string) {
			return ci.RuleDisabled, "rule disabled"
		},
	}
}
