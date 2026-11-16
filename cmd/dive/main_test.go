//go:build !windows

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// the test binary runs main() when re-executed with this set, so a test can signal a real dive process
const runMainEnv = "DIVE_TEST_RUN_MAIN"

func TestMain(m *testing.M) {
	if os.Getenv(runMainEnv) == "1" {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// stuckRegistry is a registry that never answers, so a registry pull blocks with its temp dir already made
func stuckRegistry(t *testing.T) string {
	t.Helper()
	unblock := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		<-unblock
	}))
	t.Cleanup(func() {
		close(unblock)
		srv.Close()
	})
	return "registry://" + strings.TrimPrefix(srv.URL, "http://") + "/stuck:latest"
}

// runStopped starts dive on an image whose pull never finishes, sends sig once the pull has a temp dir in TMPDIR and
// returns dive's exit code and TMPDIR.
func runStopped(t *testing.T, sig syscall.Signal) (int, string) {
	t.Helper()
	ref := stuckRegistry(t)
	tmp := t.TempDir()

	cmd := exec.Command(os.Args[0], "--ci", ref)
	cmd.Env = append(os.Environ(), runMainEnv+"=1", "TMPDIR="+tmp, "DOCKER_CONFIG="+t.TempDir())
	require.NoError(t, cmd.Start())

	// the pull blocks on the registry once its temp dir exists
	require.Eventually(t, func() bool {
		dirs, _ := filepath.Glob(filepath.Join(tmp, "dive-*", "*"))
		return len(dirs) > 0
	}, 10*time.Second, 20*time.Millisecond, "dive never started reading the image")

	require.NoError(t, cmd.Process.Signal(sig))

	err := cmd.Wait()
	var exitErr *exec.ExitError
	if err != nil {
		require.ErrorAs(t, err, &exitErr)
		require.False(t, exitErr.Sys().(syscall.WaitStatus).Signaled(), "dive was killed by %v instead of handling it", sig)
		return exitErr.ExitCode(), tmp
	}
	require.NoError(t, err)
	return 0, tmp
}

func TestMain_interruptExitsNonZero(t *testing.T) {
	code, tmp := runStopped(t, syscall.SIGINT)
	assert.Equal(t, 130, code)

	left, _ := filepath.Glob(filepath.Join(tmp, "dive-*"))
	assert.Empty(t, left, "temp dirs left behind")
}

func TestMain_terminateCleansUp(t *testing.T) {
	code, tmp := runStopped(t, syscall.SIGTERM)
	assert.Equal(t, 143, code)

	left, _ := filepath.Glob(filepath.Join(tmp, "dive-*"))
	assert.Empty(t, left, "temp dirs left behind")
}
