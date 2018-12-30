package ci

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/logrusorgru/aurora"
	"github.com/wagoodman/dive/image"
	"strconv"
)

func newGenericCiRule(key string, evaluator func(*image.AnalysisResult, string) (RuleStatus, string)) *GenericCiRule {
	return &GenericCiRule{
		key:       key,
		evaluator: evaluator,
	}
}

func (rule *GenericCiRule) Key() string {
	return rule.key
}

func (rule *GenericCiRule) Evaluate(result *image.AnalysisResult, value string) (RuleStatus, string) {
	return rule.evaluator(result, value)
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
	default:
		return aurora.Inverse("Unknown").String()
	}
}

func loadCiRules() []Rule {
	var rules = make([]Rule, 0)

	rules = append(rules, newGenericCiRule(
		"rules.lowestEfficiency",
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

	rules = append(rules, newGenericCiRule(
		"rules.highestWastedBytes",
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

	rules = append(rules, newGenericCiRule(
		"rules.highestUserWastedPercent",
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
