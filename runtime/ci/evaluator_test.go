package ci

import (
	"github.com/wagoodman/dive/dive/image/docker"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func Test_Evaluator(t *testing.T) {

	result := docker.TestAnalysisFromArchive(t, "../../.data/test-docker-image.tar")

	table := map[string]struct {
		efficiency     string
		wastedBytes    string
		wastedPercent  string
		expectedPass   bool
		expectedResult map[string]RuleStatus
	}{
		"allFail":           {"0.99", "1B", "0.01", false, map[string]RuleStatus{"lowestEfficiency": RuleFailed, "highestWastedBytes": RuleFailed, "highestUserWastedPercent": RuleFailed}},
		"allPass":           {"0.9", "50kB", "0.5", true, map[string]RuleStatus{"lowestEfficiency": RulePassed, "highestWastedBytes": RulePassed, "highestUserWastedPercent": RulePassed}},
		"allDisabled":       {"disabled", "disabled", "disabled", true, map[string]RuleStatus{"lowestEfficiency": RuleDisabled, "highestWastedBytes": RuleDisabled, "highestUserWastedPercent": RuleDisabled}},
		"misconfiguredHigh": {"1.1", "1BB", "10", false, map[string]RuleStatus{"lowestEfficiency": RuleMisconfigured, "highestWastedBytes": RuleMisconfigured, "highestUserWastedPercent": RuleMisconfigured}},
		"misconfiguredLow":  {"-9", "-1BB", "-0.1", false, map[string]RuleStatus{"lowestEfficiency": RuleMisconfigured, "highestWastedBytes": RuleMisconfigured, "highestUserWastedPercent": RuleMisconfigured}},
	}

	for name, test := range table {
		ciConfig := viper.New()
		ciConfig.SetDefault("rules.lowestEfficiency", test.efficiency)
		ciConfig.SetDefault("rules.highestWastedBytes", test.wastedBytes)
		ciConfig.SetDefault("rules.highestUserWastedPercent", test.wastedPercent)

		evaluator := NewCiEvaluator(ciConfig)

		pass := evaluator.Evaluate(result)

		if test.expectedPass != pass {
			t.Logf("Test: %s", name)
			t.Errorf("Test_Evaluator: expected pass=%v, got %v", test.expectedPass, pass)
		}

		if len(test.expectedResult) != len(evaluator.Results) {
			t.Logf("Test: %s", name)
			t.Errorf("Test_Evaluator: expected %v results, got %v", len(test.expectedResult), len(evaluator.Results))
		}

		for rule, actualResult := range evaluator.Results {
			expectedStatus := test.expectedResult[strings.TrimPrefix(rule, "rules.")]
			if expectedStatus != actualResult.status {
				t.Errorf("   %v: expected %v rule failures, got %v: %v", rule, expectedStatus, actualResult.status, actualResult)
			}
		}

	}

}
