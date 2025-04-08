package cli

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os/exec"
	"testing"
)

func Test_FetchImage(t *testing.T) {

	t.Run("fetch from docker engine", func(t *testing.T) {
		rootCmd := getTestCommand(t, "docker://busybox:1.37.0@sha256:ad9fa4d07136a83e69a54ef00102f579d04eba431932de3b0f098cc5d5948f9f")
		combined := Capture().WithStdout().WithStderr().Run(t, func() {
			require.NoError(t, rootCmd.Execute())
		})

		assert.Contains(t, combined, "Loading image")
		assert.Contains(t, combined, "Analyzing image")
		assert.Contains(t, combined, "Evaluating image")

		snaps.MatchSnapshot(t, combined)
	})

	t.Run("fetch from podman engine", func(t *testing.T) {
		if _, err := exec.LookPath("podman"); err != nil {
			t.Skip("podman not installed, skipping test")
		}

		image := "busybox:1.37.0@sha256:ad9fa4d07136a83e69a54ef00102f579d04eba431932de3b0f098cc5d5948f9f"

		// pull the image from podman first
		require.NoError(t, exec.Command("podman", "pull", image).Run())

		rootCmd := getTestCommand(t, fmt.Sprintf("podman://%s", image))
		combined := Capture().WithStdout().WithStderr().Run(t, func() {
			require.NoError(t, rootCmd.Execute())
		})

		assert.Contains(t, combined, "Loading image")
		assert.Contains(t, combined, "Analyzing image")
		assert.Contains(t, combined, "Evaluating image")

		snaps.MatchSnapshot(t, combined)
	})
}

func Test_FetchFailure(t *testing.T) {
	t.Run("nonexistent image", func(t *testing.T) {
		rootCmd := getTestCommand(t, "docker:wagoodman/nonexistent/image:tag")
		combined := Capture().WithStdout().WithStderr().Run(t, func() {
			require.ErrorContains(t, rootCmd.Execute(), "cannot load image: Error response from daemon: invalid reference format")
		})

		assert.Contains(t, combined, "Loading image")

		snaps.MatchSnapshot(t, combined)
	})

	t.Run("invalid image name", func(t *testing.T) {
		rootCmd := getTestCommand(t, "docker:///wagoodman/invalid:image:format")
		combined := Capture().WithStdout().WithStderr().Run(t, func() {
			require.ErrorContains(t, rootCmd.Execute(), "cannot load image: Error response from daemon: invalid reference format")
		})

		assert.Contains(t, combined, "Loading image")

		snaps.MatchSnapshot(t, combined)
	})
}
