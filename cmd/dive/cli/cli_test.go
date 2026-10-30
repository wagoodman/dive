package cli

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anchore/clio"
	"github.com/charmbracelet/lipgloss"
	cliconfig "github.com/docker/cli/cli/config"
	snapsPkg "github.com/gkampitakis/go-snaps/snaps"
	"github.com/google/shlex"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
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

	if _, err := snapsPkg.Clean(m); err != nil {
		fmt.Fprintln(os.Stderr, err)
		v = 1
	}

	os.Exit(v)
}

func TestUpdateSnapshotDisabled(t *testing.T) {
	require.False(t, *updateSnapshot, "update snapshot flag should be disabled")
}

// requireEngine skips the test when the given container engine CLI is missing or can't reach its daemon. The
// result is probed once per engine.
func requireEngine(t testing.TB, engine string) {
	t.Helper()
	probe, _ := engineProbes.LoadOrStore(engine, sync.OnceValue(func() error {
		if _, err := exec.LookPath(engine); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, engine, "info").CombinedOutput()
		if err != nil {
			return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
		}
		return nil
	}))
	if err := probe.(func() error)(); err != nil {
		t.Skipf("%s is not reachable, skipping: %v", engine, err)
	}
}

var engineProbes sync.Map

// requireDocker also needs the docker config to be readable: unlike the docker cli, dive won't pick a docker host
// without it
func requireDocker(t testing.TB) {
	t.Helper()
	if _, err := cliconfig.Load(""); err != nil {
		t.Skipf("docker config is not readable, skipping: %v", err)
	}
	requireEngine(t, "docker")
}

// sanitize makes output that mentions repo paths independent of where the repo is checked out
func sanitize(t testing.TB, s string) string {
	t.Helper()
	return strings.ReplaceAll(s, repoRoot(t), "<repo>")
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
	switch os.Getenv("DIVE_CONFIG") {
	case "":
		t.Setenv("DIVE_CONFIG", "./testdata/dive-enable-ci.yaml")
	case "-":
		t.Setenv("DIVE_CONFIG", "")
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
