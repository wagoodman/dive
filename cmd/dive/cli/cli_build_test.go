package cli

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func Test_Build_Dockerfile(t *testing.T) {
	requireDocker(t)
	t.Setenv("DIVE_CONFIG", "./testdata/image-multi-layer-dockerfile/dive-pass.yaml")

	t.Run("implicit dockerfile", func(t *testing.T) {
		rootCmd := getTestCommand(t, "build testdata/image-multi-layer-dockerfile")
		stdout := Capture().WithStdout().WithSuppress().Run(t, func() {
			require.NoError(t, rootCmd.Execute())
		})
		snaps.MatchSnapshot(t, stdout)
	})

	t.Run("explicit file flag", func(t *testing.T) {
		rootCmd := getTestCommand(t, "build testdata/image-multi-layer-dockerfile -f testdata/image-multi-layer-dockerfile/Dockerfile")
		stdout := Capture().WithStdout().WithSuppress().Run(t, func() {
			require.NoError(t, rootCmd.Execute())
		})
		snaps.MatchSnapshot(t, stdout)
	})
}

func Test_Build_Containerfile(t *testing.T) {
	requireDocker(t)
	t.Setenv("DIVE_CONFIG", "./testdata/image-multi-layer-containerfile/dive-pass.yaml")

	t.Run("implicit containerfile", func(t *testing.T) {
		rootCmd := getTestCommand(t, "build testdata/image-multi-layer-containerfile")
		stdout := Capture().WithStdout().WithSuppress().Run(t, func() {
			require.NoError(t, rootCmd.Execute())
		})
		snaps.MatchSnapshot(t, stdout)
	})

	t.Run("explicit file flag", func(t *testing.T) {
		rootCmd := getTestCommand(t, "build testdata/image-multi-layer-containerfile -f testdata/image-multi-layer-containerfile/Containerfile")
		stdout := Capture().WithStdout().WithSuppress().Run(t, func() {
			require.NoError(t, rootCmd.Execute())
		})
		snaps.MatchSnapshot(t, stdout)
	})
}

func Test_Build_CI_gate_fail(t *testing.T) {
	requireDocker(t)
	t.Setenv("DIVE_CONFIG", "./testdata/image-multi-layer-dockerfile/dive-fail.yaml")

	rootCmd := getTestCommand(t, "build testdata/image-multi-layer-dockerfile")
	stdout := Capture().WithStdout().WithSuppress().Run(t, func() {
		// failing gate should result in a non-zero exit code
		require.Error(t, rootCmd.Execute())
	})
	snaps.MatchSnapshot(t, stdout)

}

func Test_Build_help(t *testing.T) {
	for _, args := range []string{"build --help", "build -h", "build -t app . --help"} {
		t.Run(args, func(t *testing.T) {
			rootCmd := getTestCommand(t, args)
			combined := Capture().WithStdout().WithStderr().Run(t, func() {
				require.NoError(t, rootCmd.Execute())
			})
			assert.Contains(t, combined, "Usage:")
			assert.NotContains(t, combined, "Building image")
		})
	}
}

func Test_Build_diveFlags(t *testing.T) {
	// dive's flags come before "--", an invalid one fails before anything is built
	rootCmd := getTestCommand(t, "build --platform a/b/c/d -- -t app .")
	combined := Capture().WithStdout().WithStderr().Run(t, func() {
		require.ErrorContains(t, rootCmd.Execute(), "invalid platform")
	})
	assert.NotContains(t, combined, "Building image")

	rootCmd = getTestCommand(t, "build --not-a-dive-flag -- .")
	require.ErrorContains(t, rootCmd.Execute(), "not-a-dive-flag")

	rootCmd = getTestCommand(t, "build . -- -t app")
	require.ErrorContains(t, rootCmd.Execute(), "unexpected arguments")
}

func Test_BuildFailure(t *testing.T) {

	t.Run("nonexistent directory", func(t *testing.T) {
		requireDocker(t)
		// docker reports the missing context itself
		rootCmd := getTestCommand(t, "build ./path/does/not/exist")
		combined := Capture().WithStdout().WithStderr().Run(t, func() {
			require.ErrorContains(t, rootCmd.Execute(), "docker build failed")
		})

		assert.Contains(t, combined, "Building image")
	})

	t.Run("invalid dockerfile", func(t *testing.T) {
		requireDocker(t)
		rootCmd := getTestCommand(t, "build ./testdata/invalid")
		combined := Capture().WithStdout().WithStderr().WithSuppress().Run(t, func() {

			require.ErrorContains(t, rootCmd.Execute(), "exit status 1")
		})

		assert.Contains(t, combined, "Building image")
		// ensure we're passing through docker feedback
		assert.Contains(t, combined, "unknown instruction: INVALID")

		// replace anything starting with "docker-desktop://", like "docker-desktop://dashboard/build/desktop-linux/desktop-linux/ujdmhgkwo0sqqpopsnum3xakd"
		combined = regexp.MustCompile("docker-desktop://[^ ]+").ReplaceAllString(combined, "docker-desktop://<redacted>")

		snaps.MatchSnapshot(t, combined)
	})
}
