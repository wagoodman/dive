package ci

import (
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/image"
	"strings"
	"testing"
)

func Test_Evaluator(t *testing.T) {

	result, err := image.TestLoadDockerImageTar("../../.data/test-docker-image.tar")
	if err != nil {
		t.Fatalf("Test_Export: unable to fetch analysis: %v", err)
	}

	table := map[string]struct {
		efficiency     string
		wastedBytes    string
		wastedPercent  string
		expectedPass   bool
		expectedResult map[string]RuleStatus
	}{
		"allFail":     {"0.99", "1B", "0.01", false, map[string]RuleStatus{"lowestEfficiency": RuleFailed, "highestWastedBytes": RuleFailed, "highestUserWastedPercent": RuleFailed}},
		"allPass":     {"0.9", "50kB", "0.1", true, map[string]RuleStatus{"lowestEfficiency": RulePassed, "highestWastedBytes": RulePassed, "highestUserWastedPercent": RulePassed}},
		"allDisabled": {"disabled", "disabled", "disabled", true, map[string]RuleStatus{"lowestEfficiency": RuleDisabled, "highestWastedBytes": RuleDisabled, "highestUserWastedPercent": RuleDisabled}},
	}

	for _, test := range table {
		evaluator := NewEvaluator()

		ciConfig := viper.New()
		ciConfig.SetDefault("rules.lowestEfficiency", test.efficiency)
		ciConfig.SetDefault("rules.highestWastedBytes", test.wastedBytes)
		ciConfig.SetDefault("rules.highestUserWastedPercent", test.wastedPercent)
		evaluator.Config = ciConfig

		pass := evaluator.Evaluate(result)

		if test.expectedPass != pass {
			t.Errorf("Test_Evaluator: expected pass=%v, got %v", test.expectedPass, pass)
		}

		if len(test.expectedResult) != len(evaluator.Results) {
			t.Errorf("Test_Evaluator: expected %v results, got %v", len(test.expectedResult), len(evaluator.Results))
		}

		for rule, actualResult := range evaluator.Results {
			expectedStatus := test.expectedResult[strings.TrimPrefix(rule, "rules.")]
			if expectedStatus != actualResult.status {
				t.Errorf("   %v: expected %v rule failures, got %v", rule, expectedStatus, actualResult.status)
			}
		}

	}

}
