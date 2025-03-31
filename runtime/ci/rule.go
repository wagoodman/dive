package ci

import (
	"github.com/logrusorgru/aurora/v4"
	"github.com/wagoodman/dive/dive/image"
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

type Rule interface {
	Key() string
	Configuration() string
	Evaluate(result *image.Analysis) (RuleStatus, string)
}

type RuleStatus int

type RuleResult struct {
	status  RuleStatus
	message string
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
