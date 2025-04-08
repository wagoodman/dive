package export

import (
	"flag"
	"github.com/charmbracelet/lipgloss"
	snapsPkg "github.com/gkampitakis/go-snaps/snaps"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
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

func TestUpdateSnapshotDisabled(t *testing.T) {
	require.False(t, *updateSnapshot, "update snapshot flag should be disabled")
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
