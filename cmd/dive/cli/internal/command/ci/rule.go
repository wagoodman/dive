package ci

import (
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

func (status RuleStatus) String(f format) string {
	switch status {
	case RulePassed:
		return f.Success.Render("PASS")
	case RuleFailed:
		return f.Failure.Render("FAIL")
	case RuleWarning:
		return f.Warning.Render("WARN")
	case RuleDisabled:
		return f.Disabled.Render("SKIP")
	case RuleMisconfigured:
		return f.Warning.Render("MISCONFIGURED")
	case RuleConfigured:
		return "CONFIGURED   "
	default:
		return f.Warning.Render("Unknown")
	}
}
