package options

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCI_ciEnv(t *testing.T) {
	for _, v := range []string{"true", "True", "TRUE", "1", "yes", " true\n"} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("CI", v)
			assert.True(t, DefaultCI().Enabled)
		})
	}
	for _, v := range []string{"", "false", "False", "0", "no"} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("CI", v)
			assert.False(t, DefaultCI().Enabled)
		})
	}
}

func TestCI_PostLoad_rulesOnlyInCIMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".dive-ci")
	require.NoError(t, os.WriteFile(path, []byte("rules:\n  lowestEfficiency: nope\n"), 0o600))

	// a broken rules file is ignored outside of CI mode (TUI, --json)
	c := DefaultCI()
	c.Enabled = false
	c.ConfigPath = path
	require.NoError(t, c.PostLoad())
	require.NoError(t, c.Rules.PostLoad())

	c = DefaultCI()
	c.Enabled = true
	c.ConfigPath = path
	require.NoError(t, c.PostLoad())
	require.Error(t, c.Rules.PostLoad())
}

// loadRulesFile runs a CI mode PostLoad against a rules file with the given contents
func loadRulesFile(t *testing.T, contents string) CI {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".dive-ci")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	c := DefaultCI()
	c.Enabled = true
	c.ConfigPath = path
	require.NoError(t, c.PostLoad())
	require.NoError(t, c.Rules.PostLoad())
	return c
}

func TestCI_PostLoad_ruleFileKeys(t *testing.T) {
	c := loadRulesFile(t, `rules:
  lowest-efficiency: 0.99
  highestWastedBytes: 10MB
  highest-user-wasted-percent: disabled
  highestUserWastedPercent: 0.5
`)
	assert.Equal(t, "0.99", c.Rules.LowestEfficiencyThresholdString)
	assert.Equal(t, "10MB", c.Rules.HighestWastedBytesString)
	// snake-case wins when both spellings are given
	assert.Equal(t, "disabled", c.Rules.HighestUserWastedPercentString)
}

func TestCI_PostLoad_missingConfigPath(t *testing.T) {
	c := DefaultCI()
	c.Enabled = true
	c.ConfigPath = filepath.Join(t.TempDir(), "nope.yaml")
	require.ErrorContains(t, c.PostLoad(), "does not exist")

	// the default .dive-ci is optional
	t.Chdir(t.TempDir())
	c = DefaultCI()
	c.Enabled = true
	require.NoError(t, c.PostLoad())
}

func TestCI_PostLoad_keepsExplicitDisable(t *testing.T) {
	// CI=true only sets the default, a loaded ci=false (flag, DIVE_CI or config) has to stick
	t.Setenv("CI", "true")
	c := DefaultCI()
	c.ConfigPath = ""
	c.Enabled = false
	require.NoError(t, c.PostLoad())
	assert.False(t, c.Enabled)
}
