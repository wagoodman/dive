package cli

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Test_LoadImage(t *testing.T) {
	image := "busybox:1.37.0@sha256:ad9fa4d07136a83e69a54ef00102f579d04eba431932de3b0f098cc5d5948f9f"
	archive := repoPath(t, ".data/test-docker-image.tar")

	t.Run("from docker engine", func(t *testing.T) {
		requireDocker(t)
		runWithCombinedOutput(t, fmt.Sprintf("docker://%s", image))
	})

	t.Run("from docker engine (flag)", func(t *testing.T) {
		requireDocker(t)
		runWithCombinedOutput(t, fmt.Sprintf("--source docker %s", image))
	})

	t.Run("from podman engine", func(t *testing.T) {
		requireEngine(t, "podman")
		// pull the image from podman first
		require.NoError(t, exec.Command("podman", "pull", image).Run())

		runWithCombinedOutput(t, fmt.Sprintf("podman://%s", image))
	})

	t.Run("from podman engine (flag)", func(t *testing.T) {
		requireEngine(t, "podman")

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

// regression for #442: `dive --platform <os/arch> <multi-arch image>` analyzes that platform's image
func Test_LoadImage_platform(t *testing.T) {
	// an OCI layout holding a linux/amd64 + linux/arm64 index, each image's only file is named after its arch
	idx := mutate.IndexMediaType(empty.Index, types.OCIImageIndex)
	for _, arch := range []string{"amd64", "arm64"} {
		layer, err := crane.Layer(map[string][]byte{arch: []byte(arch)})
		require.NoError(t, err)
		img, err := mutate.AppendLayers(empty.Image, layer)
		require.NoError(t, err)
		cf, err := img.ConfigFile()
		require.NoError(t, err)
		cf = cf.DeepCopy()
		cf.OS, cf.Architecture = "linux", arch
		img, err = mutate.ConfigFile(img, cf)
		require.NoError(t, err)
		idx = mutate.AppendManifests(idx, mutate.IndexAddendum{
			Add:        img,
			Descriptor: v1.Descriptor{Platform: &v1.Platform{OS: "linux", Architecture: arch}},
		})
	}
	dir := t.TempDir()
	_, err := layout.Write(dir, idx)
	require.NoError(t, err)

	for _, arch := range []string{"amd64", "arm64"} {
		t.Run(arch, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "out.json")
			rootCmd := getTestCommand(t, fmt.Sprintf("--platform linux/%s --json %s oci-dir:%s", arch, out, dir))
			Capture().WithStdout().WithStderr().Run(t, func() {
				require.NoError(t, rootCmd.Execute())
			})

			data, err := os.ReadFile(out)
			require.NoError(t, err)
			var got struct {
				Layer []struct {
					FileList []struct {
						Path string `json:"path"`
					} `json:"fileList"`
				} `json:"layer"`
			}
			require.NoError(t, json.Unmarshal(data, &got))
			var paths []string
			for _, l := range got.Layer {
				for _, f := range l.FileList {
					paths = append(paths, f.Path)
				}
			}
			assert.Equal(t, []string{arch}, paths)
		})
	}
}

func runWithCombinedOutput(t testing.TB, cmd string) {
	t.Helper()
	rootCmd := getTestCommand(t, cmd)
	combined := Capture().WithStdout().WithStderr().Run(t, func() {
		require.NoError(t, rootCmd.Execute())
	})

	assertLoadOutput(t, sanitize(t, combined))
}

func assertLoadOutput(t testing.TB, combined string) {
	t.Helper()
	assert.Contains(t, combined, "Loading image")
	assert.Contains(t, combined, "Analyzing image")
	assert.Contains(t, combined, "Evaluating image")
	snaps.MatchSnapshot(t, combined)
}

func Test_FetchFailure(t *testing.T) {
	requireDocker(t)

	t.Run("nonexistent image", func(t *testing.T) {
		rootCmd := getTestCommand(t, "docker:wagoodman/nonexistent/image:tag")
		combined := Capture().WithStdout().WithStderr().Run(t, func() {
			require.ErrorContains(t, rootCmd.Execute(), "cannot load image")
		})

		assert.Contains(t, combined, "Loading image")

		snaps.MatchSnapshot(t, combined)
	})

	t.Run("invalid image name", func(t *testing.T) {
		rootCmd := getTestCommand(t, "docker:///wagoodman/invalid:image:format")
		combined := Capture().WithStdout().WithStderr().Run(t, func() {
			require.ErrorContains(t, rootCmd.Execute(), "cannot load image")
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
