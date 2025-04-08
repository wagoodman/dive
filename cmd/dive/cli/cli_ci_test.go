package cli

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_CI_DefaultCIConfig(t *testing.T) {
	// this lets the test harness to unset any DIVE_CONFIG env var
	t.Setenv("DIVE_CONFIG", "-")

	rootCmd := getTestCommand(t, repoPath(t, ".data/test-docker-image.tar")+" -vv")
	cd(t, "testdata/default-ci-config")
	combined := Capture().WithStdout().WithStderr().Run(t, func() {
		// failing gate should result in a non-zero exit code
		require.Error(t, rootCmd.Execute())
	})

	assert.Contains(t, combined, "lowest-efficiency: \"0.96\"", "missing lowest-efficiency rule")
	assert.Contains(t, combined, "highest-wasted-bytes: 19Mb", "missing highest-wasted-bytes rule")
	assert.Contains(t, combined, "highest-user-wasted-percent: \"0.6\"", "missing highest-user-wasted-percent rule")

	snaps.MatchSnapshot(t, combined)
}

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
