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

func TestCI_PostLoad_keepsExplicitDisable(t *testing.T) {
	// CI=true only sets the default, a loaded ci=false (flag, DIVE_CI or config) has to stick
	t.Setenv("CI", "true")
	c := DefaultCI()
	c.ConfigPath = ""
	c.Enabled = false
	require.NoError(t, c.PostLoad())
	assert.False(t, c.Enabled)
}
