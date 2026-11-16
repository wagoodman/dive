package source

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	cliconfig "github.com/docker/cli/cli/config"
	ddocker "github.com/docker/cli/cli/context/docker"
	ctxstore "github.com/docker/cli/cli/context/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/internal/imagefixture"
)

// dockerHostDaemon fakes the parts of the docker engine API that stereoscope's daemon provider uses (ping, inspect,
// pull, save). It holds one image, answered under any name, and accepts API versions 1.<minAPI> to 1.<maxAPI> the way
// the real daemon does. With missing set the image only shows up after a pull.
type dockerHostDaemon struct {
	saved          []byte
	minAPI, maxAPI int
	missing        bool

	mu       sync.Mutex
	pulled   bool
	versions []string // the API version of every versioned request
}

// newDockerHostDaemon is a docker 29 daemon (API 1.44 to 1.52) holding the oci-save fixture (3 layers)
func newDockerHostDaemon(t *testing.T) *dockerHostDaemon {
	saved, err := os.ReadFile(imagefixture.Path(t, "oci-save"))
	require.NoError(t, err)
	return &dockerHostDaemon{saved: saved, minAPI: 44, maxAPI: 52}
}

func (d *dockerHostDaemon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Api-Version", fmt.Sprintf("1.%d", d.maxAPI))
	fail := func(code int, msg string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": msg})
	}
	if r.URL.Path == "/_ping" {
		_, _ = w.Write([]byte("OK"))
		return
	}
	v, rest, ok := strings.Cut(strings.TrimPrefix(r.URL.Path, "/v"), "/")
	minor, err := strconv.Atoi(strings.TrimPrefix(v, "1."))
	if !ok || err != nil {
		fail(http.StatusNotFound, "page not found")
		return
	}
	// same wording as the daemon
	switch {
	case minor < d.minAPI:
		fail(http.StatusBadRequest, fmt.Sprintf("client version %s is too old. Minimum supported API version is 1.%d, please upgrade your client to a newer version", v, d.minAPI))
		return
	case minor > d.maxAPI:
		fail(http.StatusBadRequest, fmt.Sprintf("client version %s is too new. Maximum supported API version is 1.%d", v, d.maxAPI))
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	d.versions = append(d.versions, v)
	present := !d.missing || d.pulled
	switch {
	case r.Method == http.MethodPost && rest == "images/create":
		d.pulled = true
		_, _ = w.Write([]byte(`{"status":"Pulling fs layer","id":"abc"}` + "\n" + `{"status":"Status: Downloaded newer image"}` + "\n"))
	case !present && r.Method == http.MethodGet && strings.HasPrefix(rest, "images/"):
		fail(http.StatusNotFound, "No such image")
	case r.Method == http.MethodGet && strings.HasPrefix(rest, "images/") && strings.HasSuffix(rest, "/json"):
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"Id": "sha256:" + strings.Repeat("a", 64), "Size": len(d.saved), "Os": "linux", "Architecture": "amd64"})
	case r.Method == http.MethodGet && rest == "images/get":
		w.Header().Set("Content-Type", "application/x-tar")
		_, _ = w.Write(d.saved)
	default:
		fail(http.StatusNotFound, "page not found")
	}
}

// dockerHostServeUnix serves h on a unix socket at path
func dockerHostServeUnix(t *testing.T, path string, h http.Handler) {
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	l, err := net.Listen("unix", path)
	require.NoError(t, err)
	srv := httptest.NewUnstartedServer(h)
	_ = srv.Listener.Close()
	srv.Listener = l
	srv.Start()
	t.Cleanup(srv.Close)
}

// dockerHostShortTempDir is a temp dir short enough to hold a unix socket (macOS caps socket paths at 104 bytes,
// t.TempDir() includes the test name)
func dockerHostShortTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "dive")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

// dockerHostUseConfig makes dir the docker config dir (dive's and stereoscope's), with the given contexts (name to
// host) and current context, and clears the docker env vars so only the test's own settings apply.
func dockerHostUseConfig(t *testing.T, dir, current string, contexts map[string]string) {
	require.NoError(t, os.MkdirAll(dir, 0o700))
	if current != "" {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"currentContext":"`+current+`"}`), 0o600))
	}
	st := ctxstore.New(filepath.Join(dir, "contexts"), ctxstore.NewConfig(
		func() any { return &ddocker.EndpointMeta{} },
		ctxstore.EndpointTypeGetter(ddocker.DockerEndpoint, func() any { return &ddocker.EndpointMeta{} }),
	))
	for name, host := range contexts {
		require.NoError(t, st.CreateOrUpdate(ctxstore.Metadata{
			Name:      name,
			Endpoints: map[string]any{ddocker.DockerEndpoint: ddocker.EndpointMeta{Host: host}},
		}))
	}
	for _, k := range []string{"DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_TLS", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH", "DOCKER_API_VERSION"} {
		t.Setenv(k, "")
	}
	t.Setenv("DOCKER_CONFIG", dir)
	// cliconfig.Dir() reads DOCKER_CONFIG once per process
	prev := cliconfig.Dir()
	cliconfig.SetDir(dir)
	t.Cleanup(func() { cliconfig.SetDir(prev) })
}

// dockerHostLoad loads ref through the docker source and checks it is the fake daemon's image
func dockerHostLoad(t *testing.T, ref string) {
	img, err := Load(context.Background(), "docker://"+ref, Config{})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, img.Cleanup()) })
	assert.Len(t, img.Layers, 3)
}

// regression for #542 and #453: Docker Desktop's daemon only listens on ~/.docker/run/docker.sock, reached through
// the desktop-linux context it makes current, not on /var/run/docker.sock
func TestLoad_dockerDesktopContext(t *testing.T) {
	home := dockerHostShortTempDir(t)
	configDir := filepath.Join(home, ".docker")
	sock := filepath.Join(configDir, "run", "docker.sock")
	d := newDockerHostDaemon(t)
	dockerHostServeUnix(t, sock, d)
	dockerHostUseConfig(t, configDir, "desktop-linux", map[string]string{"desktop-linux": "unix://" + sock})

	dockerHostLoad(t, "work-database:latest")
	// the image was local to that daemon, nothing should have been pulled
	assert.False(t, d.pulled)
}

// regression for #418: with a rootless docker context current, dive pulled through the docker cli (into the rootless
// daemon) and then read from /var/run/docker.sock, which never had the image. Pull and save now go through the same
// API client, and no docker cli is needed.
func TestLoad_rootlessContextPull(t *testing.T) {
	dir := dockerHostShortTempDir(t)
	sock := filepath.Join(dir, "run", "user", "1000", "docker.sock")
	d := newDockerHostDaemon(t)
	d.missing = true
	dockerHostServeUnix(t, sock, d)
	dockerHostUseConfig(t, filepath.Join(dir, ".docker"), "rootless", map[string]string{"rootless": "unix://" + sock})
	t.Setenv("PATH", t.TempDir())

	dockerHostLoad(t, "quay.io/openshifttest/alpine:latest")
	assert.True(t, d.pulled)
}
