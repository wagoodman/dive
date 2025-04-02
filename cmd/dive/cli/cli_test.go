package cli

import (
	"bytes"
	"github.com/anchore/clio"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/google/shlex"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
)

// TODO: remove coverage from consideration of these tests (or make different test classes)

func Test_Config(t *testing.T) {
	t.Setenv("DIVE_CONFIG", "./testdata/image-multi-layer-dockerfile/dive.yaml")

	rootCmd := getTestCommand(t, "config --load")
	all := Capture().All().Run(t, func() {
		require.NoError(t, rootCmd.Execute())
	})

	snaps.MatchSnapshot(t, all)
}

func Test_LegacyRules(t *testing.T) {
	t.Setenv("DIVE_CONFIG", "./testdata/config/dive-ci-legacy.yaml")

	rootCmd := getTestCommand(t, "config --load")
	all := Capture().All().Run(t, func() {
		require.NoError(t, rootCmd.Execute())
	})

	assert.Contains(t, all, "lowest-efficiency: '0.95'", "missing lowest-efficiency legacy rule")
	assert.Contains(t, all, "highest-wasted-bytes: '20MB'", "missing highest-wasted-bytes legacy rule")
	assert.Contains(t, all, "highest-user-wasted-percent: '0.2'", "missing highest-user-wasted-percent legacy rule")
}

func Test_Build_Dockerfile(t *testing.T) {
	t.Setenv("DIVE_CONFIG", "./testdata/image-multi-layer-dockerfile/dive.yaml")

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
	t.Setenv("DIVE_CONFIG", "./testdata/image-multi-layer-containerfile/dive.yaml")

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

func getTestCommand(t testing.TB, cmd string) *cobra.Command {
	c := Command(clio.Identification{
		Name:    "dive",
		Version: "testing",
	})

	args, err := shlex.Split(cmd)
	require.NoError(t, err, "failed to parse command line %q", cmd)

	c.SetArgs(args)

	return c
}

type capturer struct {
	stdout   bool
	stderr   bool
	suppress bool
}

func Capture() *capturer {
	return &capturer{}
}

func (c *capturer) WithSuppress() *capturer {
	c.suppress = true
	return c
}

func (c *capturer) All() *capturer {
	c.stdout = true
	c.stderr = true
	return c
}

func (c *capturer) WithStdout() *capturer {
	c.stdout = true
	return c
}

func (c *capturer) WithStderr() *capturer {
	c.stderr = true
	return c
}

func (c *capturer) Run(t testing.TB, f func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		panic(err)
	}
	defer devNull.Close()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	if c.stdout {
		os.Stdout = w
	} else if c.suppress {
		os.Stdout = devNull
	}

	if c.stderr {
		os.Stderr = w
	} else if c.suppress {
		os.Stderr = devNull
	}

	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	f()
	require.NoError(t, w.Close())

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	return buf.String()
}
