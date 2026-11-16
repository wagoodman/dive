package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCI_PostLoad_ciEnv(t *testing.T) {
	for _, v := range []string{"true", "True", "TRUE", "1", "yes", " true\n"} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("CI", v)
			c := DefaultCI()
			c.ConfigPath = ""
			require.NoError(t, c.PostLoad())
			assert.True(t, c.Enabled)
		})
	}
	for _, v := range []string{"", "false", "False", "0", "no"} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("CI", v)
			c := DefaultCI()
			c.ConfigPath = ""
			require.NoError(t, c.PostLoad())
			assert.False(t, c.Enabled)
		})
	}
}
