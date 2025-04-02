package ci

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/wagoodman/dive/dive/image/docker"
)

func Test_Evaluator(t *testing.T) {
	// TODO: fix relative path to be relative to repo root instead (use a helper)
	result := docker.TestAnalysisFromArchive(t, "../../../../../../.data/test-docker-image.tar")

	validTests := []struct {
		name           string
		efficiency     string
		wastedBytes    string
		wastedPercent  string
		expectedPass   bool
		expectedResult map[string]RuleStatus
	}{
		{
			name:          "allFail",
			efficiency:    "0.99",
			wastedBytes:   "1B",
			wastedPercent: "0.01",
			expectedPass:  false,
			expectedResult: map[string]RuleStatus{
				"lowestEfficiency":         RuleFailed,
				"highestWastedBytes":       RuleFailed,
				"highestUserWastedPercent": RuleFailed,
			},
		},
		{
			name:          "allPass",
			efficiency:    "0.9",
			wastedBytes:   "50kB",
			wastedPercent: "0.5",
			expectedPass:  true,
			expectedResult: map[string]RuleStatus{
				"lowestEfficiency":         RulePassed,
				"highestWastedBytes":       RulePassed,
				"highestUserWastedPercent": RulePassed,
			},
		},
		{
			name:          "allDisabled",
			efficiency:    "disabled",
			wastedBytes:   "disabled",
			wastedPercent: "disabled",
			expectedPass:  true,
			expectedResult: map[string]RuleStatus{
				"lowestEfficiency":         RuleDisabled,
				"highestWastedBytes":       RuleDisabled,
				"highestUserWastedPercent": RuleDisabled,
			},
		},
		{
			name:          "mixedResults",
			efficiency:    "0.9",
			wastedBytes:   "1B",
			wastedPercent: "0.5",
			expectedPass:  false,
			expectedResult: map[string]RuleStatus{
				"lowestEfficiency":         RulePassed,
				"highestWastedBytes":       RuleFailed,
				"highestUserWastedPercent": RulePassed,
			},
		},
	}

	for _, test := range validTests {
		t.Run(test.name, func(t *testing.T) {
			// Create rules - these should not error
			rules, err := Rules(test.efficiency, test.wastedBytes, test.wastedPercent)
			require.NoError(t, err)

			evaluator := NewEvaluator(rules)
			eval := evaluator.Evaluate(context.TODO(), result)

			if test.expectedPass != eval.Pass {
				t.Errorf("expected pass=%v, got %v", test.expectedPass, eval.Pass)
			}

			if len(test.expectedResult) != len(evaluator.Results) {
				t.Errorf("expected %v results, got %v", len(test.expectedResult), len(evaluator.Results))
			}

			for rule, actualResult := range evaluator.Results {
				expectedStatus := test.expectedResult[rule]
				if expectedStatus != actualResult.status {
					t.Errorf("%v: expected %v rule status, got %v: %v",
						rule, expectedStatus, actualResult.status, actualResult)
				}
			}
		})
	}

}

func Test_Evaluator_Misconfigurations(t *testing.T) {
	invalidTests := []struct {
		name          string
		efficiency    string
		wastedBytes   string
		wastedPercent string
		expectError   bool
	}{
		{
			name:          "invalid_efficiency_too_high",
			efficiency:    "1.1", // fail!
			wastedBytes:   "50kB",
			wastedPercent: "0.5",
			expectError:   true,
		},
		{
			name:          "invalid_efficiency_too_low",
			efficiency:    "-0.1", // fail!
			wastedBytes:   "50kB",
			wastedPercent: "0.5",
			expectError:   true,
		},
		{
			name:          "invalid_efficiency_format",
			efficiency:    "not_a_number", // fail!
			wastedBytes:   "50kB",
			wastedPercent: "0.5",
			expectError:   true,
		},
		{
			name:          "invalid_wasted_bytes_format",
			efficiency:    "0.9",
			wastedBytes:   "not_a_size", // fail!
			wastedPercent: "0.5",
			expectError:   true,
		},
		{
			name:          "invalid_wasted_percent_high",
			efficiency:    "0.9",
			wastedBytes:   "50kB",
			wastedPercent: "1.1", // fail!
			expectError:   true,
		},
		{
			name:          "invalid_wasted_percent_low",
			efficiency:    "0.9",
			wastedBytes:   "50kB",
			wastedPercent: "-0.1", // fail!
			expectError:   true,
		},
		{
			name:          "invalid_wasted_percent_format",
			efficiency:    "0.9",
			wastedBytes:   "50kB",
			wastedPercent: "not_a_number", // fail!
			expectError:   true,
		},
	}

	for _, test := range invalidTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Rules(test.efficiency, test.wastedBytes, test.wastedPercent)
			if test.expectError {
				require.Error(t, err, "Expected an error for invalid configuration")
			} else {
				require.NoError(t, err, "Expected no error for valid configuration")
			}
		})
	}
}
