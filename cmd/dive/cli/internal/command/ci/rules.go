package ci

import (
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/wagoodman/dive/dive/v1/image"
	"strconv"
	"strings"
)

const (
	ciKeyLowestEfficiencyThreshold = "lowestEfficiency"
	ciKeyHighestWastedBytes        = "highestWastedBytes"
	ciKeyHighestUserWastedPercent  = "highestUserWastedPercent"
)

func Rules(lowerEfficiency, highestWastedBytes, highestUserWastedPercent string) ([]Rule, error) {
	var rules []Rule
	var errs []error

	lowestEfficiencyRule, err := NewLowestEfficiencyRule(lowerEfficiency)
	if err != nil {
		errs = append(errs, err)
	}
	rules = append(rules, lowestEfficiencyRule)

	highestWastedBytesRule, err := NewHighestWastedBytesRule(highestWastedBytes)
	if err != nil {
		errs = append(errs, err)
	}
	rules = append(rules, highestWastedBytesRule)

	highestUserWastedPercentRule, err := NewHighestUserWastedPercentRule(highestUserWastedPercent)
	if err != nil {
		errs = append(errs, err)
	}
	rules = append(rules, highestUserWastedPercentRule)

	return rules, errors.Join(errs...)

}

func DisabledRule(key string) Rule {
	return &BaseRule{
		key:         key,
		configValue: "disabled",
		evaluator: func(_ *image.Analysis) (RuleStatus, string) {
			return RuleDisabled, "rule disabled"
		},
	}
}

type BaseRule struct {
	key         string
	configValue string
	status      RuleStatus
	evaluator   func(*image.Analysis) (RuleStatus, string)
}

func (rule *BaseRule) Key() string {
	return rule.key
}

func (rule *BaseRule) Configuration() string {
	return rule.configValue
}

func (rule *BaseRule) Evaluate(result *image.Analysis) (RuleStatus, string) {
	if rule.status != RuleUnknown {
		return rule.status, ""
	}
	return rule.evaluator(result)
}

// LowestEfficiencyRule checks if image efficiency is above threshold
type LowestEfficiencyRule struct {
	BaseRule
	threshold float64
}

// HighestWastedBytesRule checks if wasted bytes are below threshold
type HighestWastedBytesRule struct {
	BaseRule
	threshold uint64
}

// HighestUserWastedPercentRule checks if percentage of wasted bytes is below threshold
type HighestUserWastedPercentRule struct {
	BaseRule
	threshold float64
}

func NewLowestEfficiencyRule(configValue string) (Rule, error) {
	if isRuleDisabled(configValue) {
		return DisabledRule(ciKeyLowestEfficiencyThreshold), nil
	}

	threshold, err := strconv.ParseFloat(configValue, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid %s config value, given %q: %v",
			ciKeyLowestEfficiencyThreshold, configValue, err)
	}

	if threshold < 0 || threshold > 1 {
		return nil, fmt.Errorf("%s config value is outside allowed range (0-1), given '%f'",
			ciKeyLowestEfficiencyThreshold, threshold)
	}

	return &LowestEfficiencyRule{
		BaseRule: BaseRule{
			key:         ciKeyLowestEfficiencyThreshold,
			configValue: configValue,
		},
		threshold: threshold,
	}, nil
}

func (r *LowestEfficiencyRule) Evaluate(analysis *image.Analysis) (RuleStatus, string) {
	if r.threshold > analysis.Efficiency {
		return RuleFailed, fmt.Sprintf(
			"image efficiency is too low (efficiency=%2.2f < threshold=%v)",
			analysis.Efficiency, r.threshold)
	}
	return RulePassed, ""
}

// NewHighestWastedBytesRule creates a new rule to check wasted bytes
func NewHighestWastedBytesRule(configValue string) (Rule, error) {
	if isRuleDisabled(configValue) {
		return DisabledRule(ciKeyHighestWastedBytes), nil
	}

	threshold, err := humanize.ParseBytes(configValue)
	if err != nil {
		return nil, fmt.Errorf("invalid highestWastedBytes config value, given %q: %v",
			configValue, err)
	}

	return &HighestWastedBytesRule{
		BaseRule: BaseRule{
			key:         ciKeyHighestWastedBytes,
			configValue: configValue,
		},
		threshold: threshold,
	}, nil
}

func (r *HighestWastedBytesRule) Evaluate(analysis *image.Analysis) (RuleStatus, string) {
	if analysis.WastedBytes > r.threshold {
		return RuleFailed, fmt.Sprintf(
			"too many bytes wasted (wasted-bytes=%d > threshold=%v)",
			analysis.WastedBytes, r.threshold)
	}
	return RulePassed, ""
}

// NewHighestUserWastedPercentRule creates a new rule to check percentage of wasted bytes
func NewHighestUserWastedPercentRule(configValue string) (Rule, error) {
	if isRuleDisabled(configValue) {
		return DisabledRule(ciKeyHighestUserWastedPercent), nil
	}

	threshold, err := strconv.ParseFloat(configValue, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid highestUserWastedPercent config value, given %q: %v",
			configValue, err)
	}

	if threshold < 0 || threshold > 1 {
		return nil, fmt.Errorf("highestUserWastedPercent config value is outside allowed range (0-1), given '%f'",
			threshold)
	}

	return &HighestUserWastedPercentRule{
		BaseRule: BaseRule{
			key:         ciKeyHighestUserWastedPercent,
			configValue: configValue,
		},
		threshold: threshold,
	}, nil
}

func (r *HighestUserWastedPercentRule) Evaluate(analysis *image.Analysis) (RuleStatus, string) {
	if analysis.WastedUserPercent > r.threshold {
		return RuleFailed, fmt.Sprintf(
			"too many bytes wasted, relative to the user bytes added (%%-user-wasted-bytes=%2.2f > threshold=%v)",
			analysis.WastedUserPercent, r.threshold)
	}
	return RulePassed, ""
}

func isRuleDisabled(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "" || value == "disabled" || value == "off" || value == "false"
}
