package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplication_TUIPreferences_diffHide(t *testing.T) {
	a := DefaultApplication()
	a.UI.Diff.Hide = []string{"added", "Unmodified", "bogus"}
	require.NoError(t, a.UI.Diff.PostLoad())

	// invalid values are dropped so they can't stop the TUI from starting
	assert.Equal(t, []string{"added", "Unmodified"}, a.TUIPreferences().FiletreeDiffHide)
}
