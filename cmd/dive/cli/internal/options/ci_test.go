package options

import (
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

func TestCI_PostLoad_keepsExplicitDisable(t *testing.T) {
	// CI=true only sets the default, a loaded ci=false (flag, DIVE_CI or config) has to stick
	t.Setenv("CI", "true")
	c := DefaultCI()
	c.ConfigPath = ""
	c.Enabled = false
	require.NoError(t, c.PostLoad())
	assert.False(t, c.Enabled)
}
