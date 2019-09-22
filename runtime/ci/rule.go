package ci

import (
	"fmt"
	"github.com/wagoodman/dive/dive/image"
	"strconv"

	"github.com/spf13/viper"

	"github.com/dustin/go-humanize"
	"github.com/logrusorgru/aurora"
)

const (
	RuleUnknown = iota
	RulePassed
	RuleFailed
	RuleWarning
	RuleDisabled
	RuleMisconfigured
	RuleConfigured
)

type CiRule interface {
	Key() string
	Configuration() string
	Validate() error
	Evaluate(result *image.AnalysisResult) (RuleStatus, string)
}

type GenericCiRule struct {
	key             string
	configValue     string
	configValidator func(string) error
	evaluator       func(*image.AnalysisResult, string) (RuleStatus, string)
}

type RuleStatus int

type RuleResult struct {
	status  RuleStatus
	message string
}

func newGenericCiRule(key string, configValue string, validator func(string) error, evaluator func(*image.AnalysisResult, string) (RuleStatus, string)) *GenericCiRule {
	return &GenericCiRule{
		key:             key,
		configValue:     configValue,
		configValidator: validator,
		evaluator:       evaluator,
	}
}

func (rule *GenericCiRule) Key() string {
	return rule.key
}

func (rule *GenericCiRule) Configuration() string {
	return rule.configValue
}

func (rule *GenericCiRule) Validate() error {
	return rule.configValidator(rule.configValue)
}

func (rule *GenericCiRule) Evaluate(result *image.AnalysisResult) (RuleStatus, string) {
	return rule.evaluator(result, rule.configValue)
}

func (status RuleStatus) String() string {
	switch status {
	case RulePassed:
		return "PASS"
	case RuleFailed:
		return aurora.Bold(aurora.Inverse(aurora.Red("FAIL"))).String()
	case RuleWarning:
		return aurora.Blue("WARN").String()
	case RuleDisabled:
		return aurora.Blue("SKIP").String()
	case RuleMisconfigured:
		return aurora.Bold(aurora.Inverse(aurora.Red("MISCONFIGURED"))).String()
	case RuleConfigured:
		return "CONFIGURED   "
	default:
		return aurora.Inverse("Unknown").String()
	}
}

func loadCiRules(config *viper.Viper) []CiRule {
	var rules = make([]CiRule, 0)
	var ruleKey = "lowestEfficiency"
	rules = append(rules, newGenericCiRule(
		ruleKey,
		config.GetString(fmt.Sprintf("rules.%s", ruleKey)),
		func(value string) error {
			lowestEfficiency, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid config value ('%v'): %v", value, err)
			}
			if lowestEfficiency < 0 || lowestEfficiency > 1 {
				return fmt.Errorf("lowestEfficiency config value is outside allowed range (0-1), given '%s'", value)
			}
			return nil
		},
		func(analysis *image.AnalysisResult, value string) (RuleStatus, string) {
			lowestEfficiency, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return RuleFailed, fmt.Sprintf("invalid config value ('%v'): %v", value, err)
			}
			if lowestEfficiency > analysis.Efficiency {
				return RuleFailed, fmt.Sprintf("image efficiency is too low (efficiency=%v < threshold=%v)", analysis.Efficiency, lowestEfficiency)
			}
			return RulePassed, ""
		},
	))

	ruleKey = "highestWastedBytes"
	rules = append(rules, newGenericCiRule(
		ruleKey,
		config.GetString(fmt.Sprintf("rules.%s", ruleKey)),
		func(value string) error {
			_, err := humanize.ParseBytes(value)
			if err != nil {
				return fmt.Errorf("invalid config value ('%v'): %v", value, err)
			}
			return nil
		},
		func(analysis *image.AnalysisResult, value string) (RuleStatus, string) {
			highestWastedBytes, err := humanize.ParseBytes(value)
			if err != nil {
				return RuleFailed, fmt.Sprintf("invalid config value ('%v'): %v", value, err)
			}
			if analysis.WastedBytes > highestWastedBytes {
				return RuleFailed, fmt.Sprintf("too many bytes wasted (wasted-bytes=%v > threshold=%v)", analysis.WastedBytes, highestWastedBytes)
			}
			return RulePassed, ""
		},
	))

	ruleKey = "highestUserWastedPercent"
	rules = append(rules, newGenericCiRule(
		ruleKey,
		config.GetString(fmt.Sprintf("rules.%s", ruleKey)),
		func(value string) error {
			highestUserWastedPercent, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid config value ('%v'): %v", value, err)
			}
			if highestUserWastedPercent < 0 || highestUserWastedPercent > 1 {
				return fmt.Errorf("highestUserWastedPercent config value is outside allowed range (0-1), given '%s'", value)
			}
			return nil
		},
		func(analysis *image.AnalysisResult, value string) (RuleStatus, string) {
			highestUserWastedPercent, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return RuleFailed, fmt.Sprintf("invalid config value ('%v'): %v", value, err)
			}
			if highestUserWastedPercent < analysis.WastedUserPercent {
				return RuleFailed, fmt.Sprintf("too many bytes wasted, relative to the user bytes added (%%-user-wasted-bytes=%v > threshold=%v)", analysis.WastedUserPercent, highestUserWastedPercent)
			}

			return RulePassed, ""
		},
	))

	return rules
}
