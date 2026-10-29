package ci_test

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/cmd/dive/cli/internal/command/ci"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/internal/imagefixture"
)

type ruleGolden struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type evalGolden struct {
	Pass   bool                  `json:"pass"`
	Rules  map[string]ruleGolden `json:"rules"`
	Report string                `json:"report"`
}

type ciGolden struct {
	// floats are strings since a single layer image has a NaN user wasted percent
	Efficiency        string                `json:"efficiency,omitempty"`
	SizeBytes         uint64                `json:"sizeBytes,omitempty"`
	UserSizeBytes     uint64                `json:"userSizeBytes,omitempty"`
	WastedBytes       uint64                `json:"wastedBytes,omitempty"`
	WastedUserPercent string                `json:"wastedUserPercent,omitempty"`
	Evaluations       map[string]evalGolden `json:"evaluations,omitempty"`
}

// Test_CIGoldens checks CI evaluation results of every image fixture, loaded through stereoscope, with the default
// rules and with .data/.dive-ci. Regenerate with UPDATE_GOLDEN=true.
func Test_CIGoldens(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)

	defaultRules := options.DefaultCIRules()
	require.NoError(t, defaultRules.PostLoad())

	fileCfg := options.DefaultCI()
	fileCfg.ConfigPath = imagefixture.RepoPath(".data/.dive-ci")
	require.NoError(t, fileCfg.PostLoad())
	require.NoError(t, fileCfg.Rules.PostLoad())

	ruleSets := map[string][]ci.Rule{
		"default": defaultRules.List,
		"dive-ci": fileCfg.Rules.List,
	}

	for _, name := range imagefixture.Names() {
		t.Run(name, func(t *testing.T) {
			analysis, err := dive.Analyze(context.Background(), name, imagefixture.Load(t, name))
			require.NoError(t, err)
			golden := ciGolden{
				Efficiency:        strconv.FormatFloat(analysis.Efficiency, 'g', -1, 64),
				SizeBytes:         analysis.SizeBytes,
				UserSizeBytes:     analysis.UserSizeBytes,
				WastedBytes:       analysis.WastedBytes,
				WastedUserPercent: strconv.FormatFloat(analysis.WastedUserPercent, 'g', -1, 64),
				Evaluations:       map[string]evalGolden{},
			}
			for setName, rules := range ruleSets {
				evaluator := ci.NewEvaluator(rules)
				eval := evaluator.Evaluate(context.Background(), ci.NewStats(analysis))
				g := evalGolden{Pass: eval.Pass, Rules: map[string]ruleGolden{}, Report: eval.Report}
				for key, result := range evaluator.Results {
					g.Rules[key] = ruleGolden{Status: result.StatusString(), Message: result.Message()}
				}
				golden.Evaluations[setName] = g
			}
			payload, err := json.MarshalIndent(golden, "", "  ")
			require.NoError(t, err)
			imagefixture.AssertGolden(t, imagefixture.GoldenPath(name, "ci"), append(payload, '\n'))
		})
	}
}
