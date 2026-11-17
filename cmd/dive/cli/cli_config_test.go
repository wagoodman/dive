package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Config(t *testing.T) {
	t.Setenv("DIVE_CONFIG", "./testdata/image-multi-layer-dockerfile/dive-pass.yaml")

	rootCmd := getTestCommand(t, "config --load")
	all := Capture().All().Run(t, func() {
		require.NoError(t, rootCmd.Execute())
	})

	snaps.MatchSnapshot(t, all)
}

// regression for #643: `container-engine` in the config file (not just --source) picks where a reference without a
// scheme is loaded from
func Test_Config_containerEngine(t *testing.T) {
	archive := repoPath(t, ".data/test-docker-image.tar")

	run := func(t *testing.T, engine string) error {
		cfg := filepath.Join(t.TempDir(), "dive.yaml")
		require.NoError(t, os.WriteFile(cfg, []byte("container-engine: "+engine+"\n"), 0o600))
		t.Setenv("DIVE_CONFIG", cfg)

		rootCmd := getTestCommand(t, archive+" --json "+filepath.Join(t.TempDir(), "out.json"))
		var err error
		Capture().WithStdout().WithStderr().Run(t, func() {
			err = rootCmd.Execute()
		})
		return err
	}

	t.Run("docker-archive", func(t *testing.T) {
		require.NoError(t, run(t, "docker-archive"))
	})

	// control: the same reference handed to the docker engine is not a loadable image
	t.Run("docker", func(t *testing.T) {
		assert.ErrorContains(t, run(t, "docker"), "cannot load image")
	})
}
