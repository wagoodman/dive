package cli

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_CI_Fail(t *testing.T) {
	t.Setenv("DIVE_CONFIG", "./testdata/image-multi-layer-dockerfile/dive-fail.yaml")

	rootCmd := getTestCommand(t, "build testdata/image-multi-layer-dockerfile")
	stdout := Capture().WithStdout().WithSuppress().Run(t, func() {
		// failing gate should result in a non-zero exit code
		require.Error(t, rootCmd.Execute())
	})
	snaps.MatchSnapshot(t, stdout)

}

func Test_CI_LegacyRules(t *testing.T) {
	t.Setenv("DIVE_CONFIG", "./testdata/config/dive-ci-legacy.yaml")

	rootCmd := getTestCommand(t, "config --load")
	all := Capture().All().Run(t, func() {
		require.NoError(t, rootCmd.Execute())
	})

	// this proves that we can load the legacy rules and map them to the standard rules
	assert.Contains(t, all, "lowest-efficiency: '0.95'", "missing lowest-efficiency legacy rule")
	assert.Contains(t, all, "highest-wasted-bytes: '20MB'", "missing highest-wasted-bytes legacy rule")
	assert.Contains(t, all, "highest-user-wasted-percent: '0.2'", "missing highest-user-wasted-percent legacy rule")
}
