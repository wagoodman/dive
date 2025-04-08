package cli

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func Test_JsonOutput(t *testing.T) {

	t.Run("json output", func(t *testing.T) {
		dest := t.TempDir()
		file := filepath.Join(dest, "output.json")
		rootCmd := getTestCommand(t, "busybox:1.37.0@sha256:ad9fa4d07136a83e69a54ef00102f579d04eba431932de3b0f098cc5d5948f9f --json "+file)
		combined := Capture().WithStdout().WithStderr().Run(t, func() {
			require.NoError(t, rootCmd.Execute())
		})

		assert.Contains(t, combined, "Exporting details")
		assert.Contains(t, combined, "file")

		contents, err := os.ReadFile(file)
		require.NoError(t, err)

		snaps.MatchJSON(t, contents)
	})
}
