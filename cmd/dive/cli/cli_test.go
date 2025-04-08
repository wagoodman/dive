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
	updateSnapshot = flag.Bool("update", false, "update any test snapshots")
	snaps          *snapsPkg.Config
	repoRootCache  atomic.String
)

func TestMain(m *testing.M) {
	// flags are not parsed until after test.Main is called...
	flag.Parse()

	os.Unsetenv("DIVE_CONFIG")

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
	if os.Getenv("DIVE_CONFIG") == "" {
		t.Setenv("DIVE_CONFIG", "./testdata/dive-enable-ci.yaml")
	}

	// need basic output to logger for testing...
	//l, err := logrus.New(logrus.DefaultConfig())
	//require.NoError(t, err)
	//log.Set(l)

	// get the root command
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
