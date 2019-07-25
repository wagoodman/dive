package ci

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/viper"

	"github.com/logrusorgru/aurora"
	"github.com/wagoodman/dive/image"
)

type Evaluator struct {
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

func NewEvaluator(config *viper.Viper) *Evaluator {
	return &Evaluator{
		Rules:   loadCiRules(config),
		Results: make(map[string]RuleResult),
		Pass:    true,
	}
}

func (ci *Evaluator) isRuleEnabled(rule Rule) bool {
	return rule.Configuration() != "disabled"
}

func (ci *Evaluator) Evaluate(analysis *image.AnalysisResult) bool {
	for _, rule := range ci.Rules {
		if !ci.isRuleEnabled(rule) {
			ci.Results[rule.Key()] = RuleResult{
				status:  RuleDisabled,
				message: "rule disabled",
			}
			continue
		}

		status, message := rule.Evaluate(analysis)

		if _, exists := ci.Results[rule.Key()]; exists {
			panic(fmt.Errorf("CI rule result recorded twice: %s", rule.Key()))
		}

		if status == RuleFailed {
			ci.Pass = false
		}

		ci.Results[rule.Key()] = RuleResult{
			status:  status,
			message: message,
		}

	}

	ci.Tally.Total = len(ci.Results)
	for rule, result := range ci.Results {
		switch result.status {
		case RulePassed:
			ci.Tally.Pass++
		case RuleFailed:
			ci.Tally.Fail++
		case RuleWarning:
			ci.Tally.Warn++
		case RuleDisabled:
			ci.Tally.Skip++
		default:
			panic(fmt.Errorf("unknown test status (rule='%v'): %v", rule, result.status))
		}
	}

	return ci.Pass
}

func (ci *Evaluator) Report() {
	status := "PASS"

	rules := make([]string, 0, len(ci.Results))
	for name := range ci.Results {
		rules = append(rules, name)
	}
	sort.Strings(rules)

	if ci.Tally.Fail > 0 {
		status = "FAIL"
	}

	for _, rule := range rules {
		result := ci.Results[rule]
		name := strings.TrimPrefix(rule, "rules.")
		if result.message != "" {
			fmt.Printf("  %s: %s: %s\n", result.status.String(), name, result.message)
		} else {
			fmt.Printf("  %s: %s\n", result.status.String(), name)
		}
	}

	summary := fmt.Sprintf("Result:%s [Total:%d] [Passed:%d] [Failed:%d] [Warn:%d] [Skipped:%d]", status, ci.Tally.Total, ci.Tally.Pass, ci.Tally.Fail, ci.Tally.Warn, ci.Tally.Skip)
	if ci.Pass {
		fmt.Println(aurora.Green(summary))
	} else if ci.Pass && ci.Tally.Warn > 0 {
		fmt.Println(aurora.Blue(summary))
	} else {
		fmt.Println(aurora.Red(summary))
	}
}
