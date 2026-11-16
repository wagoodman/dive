package source

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/adrg/xdg"
	"github.com/anchore/stereoscope/pkg/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// regression for #432: --source podman on an M1 mac said "unsupported platform". On darwin the podman source is the
// podman machine's API socket, which dive finds through `podman machine inspect` (faked here) and loads from.
func TestPodmanProvider_macMachineSocket(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fakes podman with a shell script")
	}
	if _, err := os.Stat("/run/podman/podman.sock"); err == nil {
		t.Skip("a real rootful podman socket would be picked before the machine socket")
	}

	// podman 5 puts the machine's API socket under $TMPDIR
	sock := filepath.Join(dockerHostShortTempDir(t), "podman", "podman-machine-default-api.sock")
	d := newDockerHostDaemon(t)
	// podman's docker compatible API
	d.minAPI, d.maxAPI = 24, 41
	dockerHostServeUnix(t, sock, d)

	bin := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(bin, "podman"), []byte("#!/bin/sh\necho '"+sock+"'\n"), 0o700))
	t.Setenv("PATH", bin)

	// nothing else may point at a socket: no CONTAINER_HOST, containers.conf or rootless socket
	t.Setenv("CONTAINER_HOST", "")
	prev := podmanConfigs
	podmanConfigs = nil
	t.Cleanup(func() { podmanConfigs = prev })
	t.Cleanup(xdg.Reload)
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	xdg.Reload()
	dockerHostUseConfig(t, t.TempDir(), "", nil)

	tmp := file.NewTempDirGenerator("dive-test")
	t.Cleanup(func() { _ = tmp.Cleanup() })
	ps, err := providers(Podman, "docker.io/library/ubuntu", "darwin", nil, tmp)
	require.NoError(t, err)
	require.Len(t, ps, 1)
	img, err := ps[0].Provide(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, img.Cleanup()) })
	assert.Len(t, img.Layers, 3)
}
