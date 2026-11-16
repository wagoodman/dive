package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CI_DefaultCIConfig(t *testing.T) {
	// this lets the test harness to unset any DIVE_CONFIG env var
	t.Setenv("DIVE_CONFIG", "-")

	// the rules come from the .dive-ci in the working dir
	rootCmd := getTestCommand(t, "docker-archive:"+repoPath(t, ".data/test-docker-image.tar")+" --ci -vv")
	cd(t, "testdata/default-ci-config")
	combined := Capture().WithStdout().WithStderr().Run(t, func() {
		require.NoError(t, rootCmd.Execute())
	})

	assert.Contains(t, combined, "lowest-efficiency: \"0.96\"", "missing lowest-efficiency rule")
	assert.Contains(t, combined, "highest-wasted-bytes: 19Mb", "missing highest-wasted-bytes rule")
	assert.Contains(t, combined, "highest-user-wasted-percent: \"0.6\"", "missing highest-user-wasted-percent rule")

	assert.Contains(t, combined, "PASS  lowestEfficiency (0.96)")
	assert.Contains(t, combined, "PASS  highestWastedBytes (19Mb)")
	assert.Contains(t, combined, "PASS  highestUserWastedPercent (0.6)")
}

func Test_CI_Fail(t *testing.T) {
	requireDocker(t)
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

func Test_CI_FlagOverridesEnv(t *testing.T) {
	t.Setenv("DIVE_CONFIG", "-")
	t.Setenv("CI", "true")

	// this gate would fail if CI mode were on
	file := filepath.Join(t.TempDir(), "output.json")
	rootCmd := getTestCommand(t, "docker-archive:"+repoPath(t, ".data/test-docker-image.tar")+" --ci=false --lowestEfficiency 0.999 --json "+file)
	combined := Capture().WithStdout().WithStderr().Run(t, func() {
		require.NoError(t, rootCmd.Execute())
	})

	assert.FileExists(t, file)
	assert.NotContains(t, combined, "lowestEfficiency")
}

func Test_CI_InvalidRulesFileOutsideCI(t *testing.T) {
	t.Setenv("DIVE_CONFIG", "-")
	t.Setenv("CI", "false")

	// a broken .dive-ci in the working dir must not matter when CI mode is off
	file := filepath.Join(t.TempDir(), "output.json")
	rootCmd := getTestCommand(t, "docker-archive:"+repoPath(t, ".data/test-docker-image.tar")+" --json "+file)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".dive-ci"), []byte("rules:\n  lowestEfficiency: nope\n"), 0o600))
	cd(t, dir)
	Capture().WithSuppress().Run(t, func() {
		require.NoError(t, rootCmd.Execute())
	})

	assert.FileExists(t, file)
}

func Test_CI_WithJsonExport(t *testing.T) {
	// exporting must not skip the gate, CI=true turns CI mode on in pipelines that also export
	file := filepath.Join(t.TempDir(), "output.json")
	rootCmd := getTestCommand(t, "docker-archive:"+repoPath(t, ".data/test-docker-image.tar")+" --ci --lowestEfficiency 0.999 --json "+file)
	combined := Capture().WithStdout().WithStderr().Run(t, func() {
		require.ErrorContains(t, rootCmd.Execute(), "evaluation failed")
	})

	assert.FileExists(t, file)
	assert.Contains(t, combined, "FAIL  lowestEfficiency")
}
