package cli

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"testing"
)

func Test_LoadImage(t *testing.T) {
	image := "busybox:1.37.0@sha256:ad9fa4d07136a83e69a54ef00102f579d04eba431932de3b0f098cc5d5948f9f"
	archive := repoPath(t, ".data/test-docker-image.tar")

	t.Run("from docker engine", func(t *testing.T) {
		runWithCombinedOutput(t, fmt.Sprintf("docker://%s", image))
	})

	t.Run("from docker engine (flag)", func(t *testing.T) {

		runWithCombinedOutput(t, fmt.Sprintf("--source docker %s", image))
	})

	t.Run("from podman engine", func(t *testing.T) {
		if _, err := exec.LookPath("podman"); err != nil {
			t.Skip("podman not installed, skipping test")
		}
		// pull the image from podman first
		require.NoError(t, exec.Command("podman", "pull", image).Run())

		runWithCombinedOutput(t, fmt.Sprintf("podman://%s", image))
	})

	t.Run("from podman engine (flag)", func(t *testing.T) {
		if _, err := exec.LookPath("podman"); err != nil {
			t.Skip("podman not installed, skipping test")
		}

		// pull the image from podman first
		require.NoError(t, exec.Command("podman", "pull", image).Run())

		runWithCombinedOutput(t, fmt.Sprintf("--source podman %s", image))
	})

	t.Run("from archive", func(t *testing.T) {
		runWithCombinedOutput(t, fmt.Sprintf("docker-archive://%s", archive))
	})

	t.Run("from archive (flag)", func(t *testing.T) {
		runWithCombinedOutput(t, fmt.Sprintf("--source docker-archive %s", archive))
	})
}

func runWithCombinedOutput(t testing.TB, cmd string) {
	t.Helper()
	rootCmd := getTestCommand(t, cmd)
	combined := Capture().WithStdout().WithStderr().Run(t, func() {
		require.NoError(t, rootCmd.Execute())
	})

	assertLoadOutput(t, combined)
}

func assertLoadOutput(t testing.TB, combined string) {
	t.Helper()
	assert.Contains(t, combined, "Loading image")
	assert.Contains(t, combined, "Analyzing image")
	assert.Contains(t, combined, "Evaluating image")
	snaps.MatchSnapshot(t, combined)
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

func cd(t testing.TB, to string) {
	t.Helper()
	from, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(to))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(from))
	})
}
