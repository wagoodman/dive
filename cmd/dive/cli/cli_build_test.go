package cli

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func Test_Build_Dockerfile(t *testing.T) {
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
	t.Setenv("DIVE_CONFIG", "./testdata/image-multi-layer-dockerfile/dive-fail.yaml")

	rootCmd := getTestCommand(t, "build testdata/image-multi-layer-dockerfile")
	stdout := Capture().WithStdout().WithSuppress().Run(t, func() {
		// failing gate should result in a non-zero exit code
		require.Error(t, rootCmd.Execute())
	})
	snaps.MatchSnapshot(t, stdout)

}

func Test_BuildFailure(t *testing.T) {

	t.Run("nonexistent directory", func(t *testing.T) {
		rootCmd := getTestCommand(t, "build ./path/does/not/exist")
		combined := Capture().WithStdout().WithStderr().Run(t, func() {
			require.ErrorContains(t, rootCmd.Execute(), "could not find Containerfile or Dockerfile")
		})

		assert.Contains(t, combined, "Building image")

		snaps.MatchSnapshot(t, combined)
	})

	t.Run("invalid dockerfile", func(t *testing.T) {
		rootCmd := getTestCommand(t, "build ./testdata/invalid")
		combined := Capture().WithStdout().WithStderr().WithSuppress().Run(t, func() {

			require.ErrorContains(t, rootCmd.Execute(), "cannot build image: exit status 1")
		})

		assert.Contains(t, combined, "Building image")
		// ensure we're passing through docker feedback
		assert.Contains(t, combined, "unknown instruction: INVALID")

		// replace anything starting with "docker-desktop://", like "docker-desktop://dashboard/build/desktop-linux/desktop-linux/ujdmhgkwo0sqqpopsnum3xakd"
		combined = regexp.MustCompile("docker-desktop://[^ ]+").ReplaceAllString(combined, "docker-desktop://<redacted>")

		snaps.MatchSnapshot(t, combined)
	})
}
