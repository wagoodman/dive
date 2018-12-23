package ci

import (
	"bytes"
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/image"
	"io/ioutil"
	"strings"
)

func NewEvaluator(configFile string) *Evaluator {
	ciConfig := viper.New()
	ciConfig.SetConfigType("yaml")

	ciConfig.SetDefault("rules.lowestEfficiency", 0.9)
	ciConfig.SetDefault("rules.highestWastedBytes", "disabled")
	ciConfig.SetDefault("rules.highestUserWastedPercent", 0.1)

	fileBytes, err := ioutil.ReadFile(configFile)
	if err == nil {
		fmt.Printf("  Using CI config: %s\n", configFile)
		err = ciConfig.ReadConfig(bytes.NewBuffer(fileBytes))
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("  Using default CI config")
	}

	return &Evaluator{
		Config:  ciConfig,
		Rules:   loadCiRules(),
		Results: make(map[string]RuleResult),
		Pass:    true,
	}
}

func (ci *Evaluator) isRuleEnabled(rule Rule) bool {
	value := ci.Config.GetString(rule.Key())
	if value == "disabled" {
		return false
	}
	return true
}

func (ci *Evaluator) Evaluate(analysis *image.AnalysisResult) {
	for _, rule := range ci.Rules {
		if ci.isRuleEnabled(rule) {

			value := ci.Config.GetString(rule.Key())
			status, message := rule.Evaluate(analysis, value)

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
		} else {
			ci.Results[rule.Key()] = RuleResult{
				status:  RuleDisabled,
				message: "skipped (disabled)",
			}
		}
	}
}

func (ci *Evaluator) Report() {
	numRules := len(ci.Rules)
	numPass := 0
	numFail := 0
	numSkip := 0
	numWarn := 0
	status := "PASS"
	for rule, result := range ci.Results {
		name := strings.TrimPrefix(rule, "rules.")
		switch result.status {
		case RulePassed:
			numPass++
		case RuleFailed:
			numFail++
			status = "FAIL"
		case RuleWarning:
			numWarn++
		case RuleDisabled:
			numSkip++
		default:
			panic(fmt.Errorf("unknown test status: %v", result.status))
		}
		if result.message != "" {
			fmt.Printf("  %s: %s: %s\n", result.status.String(), name, result.message)
		} else {
			fmt.Printf("  %s: %s\n", result.status.String(), name)
		}
	}

	summary := fmt.Sprintf("Result:%s [Total:%d] [Passed:%d] [Failed:%d] [Warn:%d] [Skipped:%d]", status, numRules, numPass, numFail, numWarn, numSkip)
	if ci.Pass {
		fmt.Println(aurora.Green(summary))
	} else if ci.Pass && numWarn > 0 {
		fmt.Println(aurora.Blue(summary))
	} else {
		fmt.Println(aurora.Red(summary))
	}
}
