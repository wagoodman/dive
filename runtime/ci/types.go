package ci

import (
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/image"
)

type RuleStatus int

type RuleResult struct {
	status  RuleStatus
	message string
}

const (
	RuleUnknown = iota
	RulePassed
	RuleFailed
	RuleWarning
	RuleDisabled
)

type Rule interface {
	Key() string
	Evaluate(*image.AnalysisResult, string) (RuleStatus, string)
}

type GenericCiRule struct {
	key       string
	evaluator func(*image.AnalysisResult, string) (RuleStatus, string)
}

type Evaluator struct {
	Config  *viper.Viper
	Rules   []Rule
	Results map[string]RuleResult
	Tally   ResultTally
	Pass    bool
}

type ResultTally struct {
	Pass  int
	Fail  int
	Skip  int
	Warn  int
	Total int
}
