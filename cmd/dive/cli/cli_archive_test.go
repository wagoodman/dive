package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/internal/imagefixture"
)

// these run the real command against image archives, so the CLI path is covered without a container engine

var archiveFixtures = []struct {
	fixture, scheme string
}{
	{fixture: "test-docker-image", scheme: "docker-archive:"},
	{fixture: "test-oci-docker-image", scheme: "docker-archive://"},
	{fixture: "oci-archive-gzip", scheme: "oci-archive:"},
	{fixture: "opaque-whiteout-dir", scheme: "docker-archive:"},
}

func Test_Archive_JsonOutput(t *testing.T) {
	for _, tt := range archiveFixtures {
		t.Run(tt.fixture, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "output.json")
			// a longer file at the destination must not leave trailing bytes behind
			require.NoError(t, os.WriteFile(file, make([]byte, 1<<20), 0o644))

			rootCmd := getTestCommand(t, tt.scheme+imagefixture.Path(t, tt.fixture)+" --json "+file)
			combined := Capture().WithStdout().WithStderr().Run(t, func() {
				require.NoError(t, rootCmd.Execute())
			})
			assert.Contains(t, combined, "Exporting details")

			got, err := os.ReadFile(file)
			require.NoError(t, err)
			want, err := os.ReadFile(imagefixture.GoldenPath(tt.fixture, "export"))
			require.NoError(t, err)
			assert.JSONEq(t, string(want), string(got))
		})
	}
}

func Test_Archive_CI(t *testing.T) {
	archive := "docker-archive:" + repoPath(t, ".data/test-docker-image.tar")

	t.Run("fail", func(t *testing.T) {
		t.Setenv("DIVE_CONFIG", "./testdata/image-multi-layer-dockerfile/dive-fail.yaml")
		rootCmd := getTestCommand(t, archive)
		stdout := Capture().WithStdout().WithSuppress().Run(t, func() {
			require.ErrorContains(t, rootCmd.Execute(), "evaluation failed")
		})
		snaps.MatchSnapshot(t, stdout)
	})

	t.Run("oci-archive", func(t *testing.T) {
		rootCmd := getTestCommand(t, "oci-archive:"+imagefixture.Path(t, "oci-archive-gzip"))
		stdout := Capture().WithStdout().WithSuppress().Run(t, func() {
			require.NoError(t, rootCmd.Execute())
		})
		snaps.MatchSnapshot(t, stdout)
	})

	t.Run("load failure", func(t *testing.T) {
		rootCmd := getTestCommand(t, "docker-archive:"+filepath.Join(t.TempDir(), "missing.tar"))
		combined := Capture().WithStdout().WithStderr().Run(t, func() {
			require.ErrorContains(t, rootCmd.Execute(), "cannot load image")
		})
		assert.Contains(t, combined, "Loading image")
	})
}
