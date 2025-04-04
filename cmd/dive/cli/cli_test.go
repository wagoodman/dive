package cli

import (
	"bytes"
	"flag"
	"github.com/anchore/clio"
	"github.com/charmbracelet/lipgloss"
	snapsPkg "github.com/gkampitakis/go-snaps/snaps"
	"github.com/google/shlex"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	updateSnapshot = flag.Bool("update", false, "update snapshots flag")
	snaps          *snapsPkg.Config
	repoRootCache  atomic.String
)

func TestMain(m *testing.M) {
	// flags are not parsed until after test.Main is called...
	flag.Parse()

	// disable colors
	lipgloss.SetColorProfile(termenv.Ascii)

	snaps = snapsPkg.WithConfig(
		snapsPkg.Update(*updateSnapshot),
		snapsPkg.Dir("testdata/snapshots"),
	)

	v := m.Run()

	snapsPkg.Clean(m)

	os.Exit(v)
}

func Test_Config(t *testing.T) {
	t.Setenv("DIVE_CONFIG", "./testdata/image-multi-layer-dockerfile/dive-pass.yaml")

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

	// this proves that we can load the legacy rules and map them to the standard rules
	assert.Contains(t, all, "lowest-efficiency: '0.95'", "missing lowest-efficiency legacy rule")
	assert.Contains(t, all, "highest-wasted-bytes: '20MB'", "missing highest-wasted-bytes legacy rule")
	assert.Contains(t, all, "highest-user-wasted-percent: '0.2'", "missing highest-user-wasted-percent legacy rule")
}

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

func Test_root_CI_gate_fail(t *testing.T) {
	t.Setenv("DIVE_CONFIG", "./testdata/image-multi-layer-dockerfile/dive-fail.yaml")

	rootCmd := getTestCommand(t, "build testdata/image-multi-layer-dockerfile")
	stdout := Capture().WithStdout().WithSuppress().Run(t, func() {
		// failing gate should result in a non-zero exit code
		require.Error(t, rootCmd.Execute())
	})
	snaps.MatchSnapshot(t, stdout)

}

func repoPath(t testing.TB, path string) string {
	t.Helper()
	root := repoRoot(t)
	return filepath.Join(root, path)
}

func repoRoot(t testing.TB) string {
	val := repoRootCache.Load()
	if val != "" {
		return val
	}
	t.Helper()
	// use git to find the root of the repo
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatalf("failed to get repo root: %v", err)
	}
	val = strings.TrimSpace(string(out))
	repoRootCache.Store(val)
	return val
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
