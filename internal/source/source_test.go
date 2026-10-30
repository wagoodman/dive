package source

import (
	"context"
	"encoding/pem"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anchore/stereoscope/pkg/file"
	cliconfig "github.com/docker/cli/cli/config"
	ddocker "github.com/docker/cli/cli/context/docker"
	ctxstore "github.com/docker/cli/cli/context/store"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/internal/imagefixture"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input, defaultSource string
		wantSource, wantRef  string
	}{
		// no scheme: the default source, input untouched
		{"alpine:3.20", "", Docker, "alpine:3.20"},
		{"alpine:3.20", Docker, Docker, "alpine:3.20"},
		{"alpine:3.20", Podman, Podman, "alpine:3.20"},
		{"alpine:3.20", Auto, Auto, "alpine:3.20"},
		{"img.tar", DockerArchive, DockerArchive, "img.tar"},
		{"img.tar", "docker-tar", DockerArchive, "img.tar"},
		{"img.tar", OCIArchive, OCIArchive, "img.tar"},
		{"./layout", OCIDir, OCIDir, "./layout"},
		{"alpine", Registry, Registry, "alpine"},
		{"alpine", "oci-registry", Registry, "alpine"},
		{"alpine", Containerd, Containerd, "alpine"},
		{"alpine", ContainersStorage, ContainersStorage, "alpine"},
		// unknown defaults pass through so Load can report them
		{"alpine", "nope", "nope", "alpine"},

		// dive "://" schemes beat the default
		{"docker://alpine", Podman, Docker, "alpine"},
		{"podman://alpine", Docker, Podman, "alpine"},
		{"docker-archive://img.tar", Docker, DockerArchive, "img.tar"},
		{"docker-tar://img.tar", Docker, DockerArchive, "img.tar"},
		{"oci-archive://img.tar", Docker, OCIArchive, "img.tar"},
		{"oci-dir:///abs/layout", Docker, OCIDir, "/abs/layout"},
		{"registry://alpine", Docker, Registry, "alpine"},
		{"oci-registry://alpine", Docker, Registry, "alpine"},
		{"containerd://alpine", Docker, Containerd, "alpine"},
		{"containers-storage://alpine", Docker, ContainersStorage, "alpine"},
		{"registry://localhost:5000/img:tag", Docker, Registry, "localhost:5000/img:tag"},

		// stereoscope ":" schemes, for the file based sources only
		{"docker-archive:img.tar", Docker, DockerArchive, "img.tar"},
		{"docker-tar:img.tar", Docker, DockerArchive, "img.tar"},
		{"oci-archive:img.tar", Docker, OCIArchive, "img.tar"},
		{"oci-dir:/abs/layout", Docker, OCIDir, "/abs/layout"},

		// any other source name before a ":" is an image named like that source, with a tag
		{"docker:dind", Podman, Podman, "docker:dind"},
		{"docker:27-cli", Docker, Docker, "docker:27-cli"},
		{"registry:2.8", Docker, Docker, "registry:2.8"},
		{"registry:latest", Docker, Docker, "registry:latest"},
		{"podman:alpine", Docker, Docker, "podman:alpine"},
		{"containerd:alpine", Docker, Docker, "containerd:alpine"},
		{"oci-registry:alpine", Docker, Docker, "oci-registry:alpine"},
		{"containers-storage:latest", Podman, Podman, "containers-storage:latest"},

		// "auto" is a --source value, not a scheme
		{"auto://alpine", Docker, Docker, "auto://alpine"},
		{"auto:alpine", Docker, Docker, "auto:alpine"},

		// registry hosts with a port are not schemes, even when the host is named like a source
		{"localhost:5000/img:tag", Docker, Docker, "localhost:5000/img:tag"},
		{"localhost:5000/img:tag", Registry, Registry, "localhost:5000/img:tag"},
		{"registry:5000/img:tag", Docker, Docker, "registry:5000/img:tag"},
		{"docker:5000/img", Podman, Podman, "docker:5000/img"},
		{"registry:5000", Docker, Docker, "registry:5000"},
		{"myhost.io:443/team/img@sha256:abc", Docker, Docker, "myhost.io:443/team/img@sha256:abc"},

		// unknown schemes are part of the reference
		{"https://example.com/img", Docker, Docker, "https://example.com/img"},
		{"sha256:9f5a0b6c", Docker, Docker, "sha256:9f5a0b6c"},
		{"docker://sha256:9f5a0b6c", Podman, Docker, "sha256:9f5a0b6c"},
		{`C:\images\img.tar`, DockerArchive, DockerArchive, `C:\images\img.tar`},
	}
	for _, tt := range tests {
		t.Run(tt.input+" default="+tt.defaultSource, func(t *testing.T) {
			src, ref := Parse(tt.input, tt.defaultSource)
			assert.Equal(t, tt.wantSource, src)
			assert.Equal(t, tt.wantRef, ref)
		})
	}
}

func TestNormalize(t *testing.T) {
	for _, s := range Sources {
		got, ok := Normalize(s)
		assert.True(t, ok, s)
		assert.Equal(t, s, got)
	}
	got, ok := Normalize("docker-tar")
	assert.True(t, ok)
	assert.Equal(t, DockerArchive, got)
	got, ok = Normalize("oci-registry")
	assert.True(t, ok)
	assert.Equal(t, Registry, got)
	_, ok = Normalize("nope")
	assert.False(t, ok)
	_, ok = Normalize("")
	assert.False(t, ok)
}

func TestProviders(t *testing.T) {
	tests := []struct {
		src, goos string
		want      []string
		wantErr   bool
	}{
		{src: Docker, goos: "linux", want: []string{"docker"}},
		// the short name is also tried as podman's localhost/ name
		{src: Podman, goos: "linux", want: []string{"containers-storage", "containers-storage", "podman"}},
		{src: Podman, goos: "darwin", want: []string{"podman"}},
		{src: Podman, goos: "windows", want: []string{"podman"}},
		{src: DockerArchive, goos: "linux", want: []string{"docker-archive", "oci-archive"}},
		{src: OCIArchive, goos: "linux", want: []string{"oci-archive"}},
		{src: OCIDir, goos: "linux", want: []string{"oci-dir"}},
		{src: Registry, goos: "linux", want: []string{"oci-registry"}},
		{src: Containerd, goos: "linux", want: []string{"containerd"}},
		{src: ContainersStorage, goos: "darwin", want: []string{"containers-storage"}},
		{src: Auto, goos: "linux", want: []string{
			"docker-archive", "oci-archive", "oci-dir", "docker", "podman", "containerd", "containers-storage", "oci-registry",
		}},
		{src: "nope", goos: "linux", wantErr: true},
		{src: "", goos: "linux", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.src+"/"+tt.goos, func(t *testing.T) {
			ps, err := providers(tt.src, "alpine", tt.goos, nil, file.NewTempDirGenerator("dive-test"))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			var names []string
			for _, p := range ps {
				names = append(names, p.Name())
			}
			assert.Equal(t, tt.want, names)
		})
	}
}

func TestLoad(t *testing.T) {
	dockerSave := imagefixture.Path(t, "oci-save")
	ociArchive := imagefixture.Path(t, "oci-archive-gzip")

	tests := []struct {
		name       string
		input      string
		cfg        Config
		wantLayers int
		wantErr    bool
	}{
		{name: "docker-archive scheme", input: "docker-archive://" + dockerSave, wantLayers: 3},
		{name: "docker-tar default", input: dockerSave, cfg: Config{Source: "docker-tar"}, wantLayers: 3},
		{name: "oci-archive stereoscope scheme", input: "oci-archive:" + ociArchive, wantLayers: 1},
		{name: "docker-archive reads an OCI archive", input: ociArchive, cfg: Config{Source: DockerArchive}, wantLayers: 1},
		{name: "auto detects an archive", input: ociArchive, cfg: Config{Source: Auto}, wantLayers: 1},
		{name: "missing archive", input: "docker-archive://" + filepath.Join(t.TempDir(), "nope.tar"), wantErr: true},
		{name: "unknown source", input: dockerSave, cfg: Config{Source: "nope"}, wantErr: true},
		{name: "bad platform", input: dockerSave, cfg: Config{Source: DockerArchive, Platform: "a/b/c/d"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := Load(context.Background(), tt.input, tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, img)
				return
			}
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, img.Cleanup()) })
			assert.Len(t, img.Layers, tt.wantLayers)
		})
	}
}

func TestLoad_interrupted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// auto has a provider list to walk. A cancelled load stops before the first, so no provider reports an error
	// (and nothing is pulled)
	img, err := Load(ctx, "dive-test/does-not-exist:never", Config{Source: Auto})
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, img)
	ps, perr := providers(Auto, "alpine", "linux", nil, file.NewTempDirGenerator("dive-test"))
	require.NoError(t, perr)
	for _, p := range ps {
		assert.NotContains(t, err.Error(), p.Name()+":")
	}
}

func TestProviders_plainHTTPRegistry(t *testing.T) {
	tmp := file.NewTempDirGenerator("dive-test")
	for _, ref := range []string{"10.0.5.20:5000/app", "192.168.1.2/app:1", "172.16.0.9:443/team/app", "build.localhost:5000/app", "build.localhost/app"} {
		_, err := providers(Registry, ref, "linux", nil, tmp)
		assert.ErrorContains(t, err, "plain http", ref)

		// auto never gets as far as a registry pull for these
		ps, err := providers(Auto, ref, "linux", nil, tmp)
		require.NoError(t, err)
		assert.NotEqual(t, "oci-registry", ps[len(ps)-1].Name(), ref)
	}
	for _, ref := range []string{"alpine", "ghcr.io/org/app:1", "localhost:5000/app", "127.0.0.1:5000/app", "[::1]:5000/app", "[::1]/app"} {
		_, err := providers(Registry, ref, "linux", nil, tmp)
		assert.NoError(t, err, ref)
	}
}

func TestLocalName(t *testing.T) {
	tests := []struct {
		ref, want string
		ok        bool
	}{
		{"myapp", "localhost/myapp", true},
		{"myapp:1.0", "localhost/myapp:1.0", true},
		{"team/myapp", "localhost/team/myapp", true},
		{"localhost/myapp", "", false},
		{"quay.io/team/myapp", "", false},
		{"registry:5000/myapp", "", false},
	}
	for _, tt := range tests {
		got, ok := localName(tt.ref)
		assert.Equal(t, tt.ok, ok, tt.ref)
		assert.Equal(t, tt.want, got, tt.ref)
	}
}

func TestCleanup(t *testing.T) {
	path := imagefixture.Path(t, "oci-save-gzip")
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp) // unix
	t.Setenv("TMP", tmp)    // windows

	// an image whose owner never got to img.Cleanup() (an interrupt) still gets removed
	img, err := Load(context.Background(), "docker-archive:"+path, Config{})
	require.NoError(t, err)
	require.NotEmpty(t, img.Layers)
	entries, err := os.ReadDir(tmp)
	require.NoError(t, err)
	require.NotEmpty(t, entries, "expected the image to hold temp dirs")

	require.NoError(t, Cleanup(time.Second))
	entries, err = os.ReadDir(tmp)
	require.NoError(t, err)
	assert.Empty(t, entries)

	// the owner cleaning up afterwards is fine
	require.NoError(t, img.Cleanup())
}

func TestCleanup_waitsForLoad(t *testing.T) {
	// stands in for a Load still running
	loading.Add(1)
	done := make(chan struct{})
	go func() {
		assert.NoError(t, Cleanup(10*time.Second))
		close(done)
	}()
	select {
	case <-done:
		t.Fatal("Cleanup returned while a load was still running")
	case <-time.After(100 * time.Millisecond):
	}
	loading.Done()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Cleanup did not return after the load finished")
	}

	// a load that never returns only holds Cleanup up for the wait
	loading.Add(1)
	defer loading.Done()
	start := time.Now()
	require.NoError(t, Cleanup(50*time.Millisecond))
	assert.Less(t, time.Since(start), 5*time.Second)
}

func TestDockerHost(t *testing.T) {
	newConfigDir := func(t *testing.T, currentContext string, contexts map[string]string) string {
		t.Helper()
		dir := t.TempDir()
		if currentContext != "" {
			require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"currentContext":"`+currentContext+`"}`), 0o600))
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
		return dir
	}

	contexts := map[string]string{
		"remote":  "tcp://10.0.0.1:2376",
		"ssh":     "ssh://user@builder",
		"colima":  "unix:///Users/me/.colima/default/docker.sock",
		"no-host": "",
	}

	tests := []struct {
		name           string
		dockerHost     string
		dockerContext  string
		currentContext string
		want           string
		wantErr        bool
	}{
		{name: "nothing set", want: client.DefaultDockerHost},
		{name: "DOCKER_HOST", dockerHost: "tcp://1.2.3.4:2375", want: "tcp://1.2.3.4:2375"},
		{name: "DOCKER_HOST beats DOCKER_CONTEXT", dockerHost: "tcp://1.2.3.4:2375", dockerContext: "remote", want: "tcp://1.2.3.4:2375"},
		{name: "DOCKER_HOST beats config context", dockerHost: "tcp://1.2.3.4:2375", currentContext: "remote", want: "tcp://1.2.3.4:2375"},
		{name: "DOCKER_CONTEXT", dockerContext: "remote", want: "tcp://10.0.0.1:2376"},
		{name: "DOCKER_CONTEXT ssh", dockerContext: "ssh", want: "ssh://user@builder"},
		{name: "DOCKER_CONTEXT beats config context", dockerContext: "colima", currentContext: "remote", want: "unix:///Users/me/.colima/default/docker.sock"},
		{name: "config context", currentContext: "colima", want: "unix:///Users/me/.colima/default/docker.sock"},
		{name: "context with empty host", currentContext: "no-host", want: client.DefaultDockerHost},
		{name: "DOCKER_CONTEXT default", dockerContext: "default", currentContext: "remote", want: client.DefaultDockerHost},
		{name: "config context default", currentContext: "default", want: client.DefaultDockerHost},
		{name: "missing context", dockerContext: "missing", wantErr: true},
		{name: "missing config context", currentContext: "missing", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DOCKER_HOST", tt.dockerHost)
			t.Setenv("DOCKER_CONTEXT", tt.dockerContext)
			ep, _, err := dockerEndpoint(newConfigDir(t, tt.currentContext, contexts))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, ep.Host)
		})
	}
}

func TestDockerHost_badConfig(t *testing.T) {
	t.Setenv("DOCKER_HOST", "")
	t.Setenv("DOCKER_CONTEXT", "")
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte("{not json"), 0o600))
	_, _, err := dockerEndpoint(dir)
	require.Error(t, err)
}

func TestNewDockerClient_TLS(t *testing.T) {
	// a TLS only daemon, a client that skips TLS gets a 400 from it
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Api-Version", "1.44")
		w.WriteHeader(http.StatusOK)
	}))
	srv.Config.ErrorLog = log.New(io.Discard, "", 0)
	srv.StartTLS()
	defer srv.Close()
	host := "tcp://" + srv.Listener.Addr().String()
	ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})

	newConfigDir := func(t *testing.T, withCA bool) string {
		dir := t.TempDir()
		st := ctxstore.New(filepath.Join(dir, "contexts"), ctxstore.NewConfig(
			func() any { return &ddocker.EndpointMeta{} },
			ctxstore.EndpointTypeGetter(ddocker.DockerEndpoint, func() any { return &ddocker.EndpointMeta{} }),
		))
		require.NoError(t, st.CreateOrUpdate(ctxstore.Metadata{
			Name:      "secure",
			Endpoints: map[string]any{ddocker.DockerEndpoint: ddocker.EndpointMeta{Host: host}},
		}))
		require.NoError(t, st.CreateOrUpdate(ctxstore.Metadata{
			Name:      "plain",
			Endpoints: map[string]any{ddocker.DockerEndpoint: ddocker.EndpointMeta{Host: host}},
		}))
		require.NoError(t, st.ResetTLSMaterial("secure", &ctxstore.ContextTLSData{
			Endpoints: map[string]ctxstore.EndpointTLSData{ddocker.DockerEndpoint: {Files: map[string][]byte{"ca.pem": ca}}},
		}))
		if withCA {
			require.NoError(t, os.WriteFile(filepath.Join(dir, "ca.pem"), ca, 0o600))
		}
		return dir
	}

	tests := []struct {
		name    string
		env     map[string]string
		withCA  bool
		wantErr bool
	}{
		{name: "context TLS material", env: map[string]string{"DOCKER_CONTEXT": "secure"}},
		{name: "context without TLS material", env: map[string]string{"DOCKER_CONTEXT": "plain"}, wantErr: true},
		{name: "DOCKER_TLS_VERIFY with config dir CA", env: map[string]string{"DOCKER_HOST": host, "DOCKER_TLS_VERIFY": "1"}, withCA: true},
		{name: "DOCKER_TLS_VERIFY without a CA", env: map[string]string{"DOCKER_HOST": host, "DOCKER_TLS_VERIFY": "1"}, wantErr: true},
		{name: "DOCKER_TLS unverified", env: map[string]string{"DOCKER_HOST": host, "DOCKER_TLS": "1"}},
		{name: "DOCKER_HOST without TLS", env: map[string]string{"DOCKER_HOST": host}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range []string{"DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_TLS", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH"} {
				t.Setenv(k, tt.env[k])
			}
			// cliconfig.Dir() reads DOCKER_CONFIG once per process
			prev := cliconfig.Dir()
			cliconfig.SetDir(newConfigDir(t, tt.withCA))
			t.Cleanup(func() { cliconfig.SetDir(prev) })

			c, err := newDockerClient()
			if err == nil {
				defer c.Close()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_, err = c.Ping(ctx, client.PingOptions{})
			}
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPodmanHost(t *testing.T) {
	writeConf := func(t *testing.T, content string) string {
		p := filepath.Join(t.TempDir(), "containers.conf")
		require.NoError(t, os.WriteFile(p, []byte(content), 0o600))
		return p
	}
	sshOnly := writeConf(t, `[engine]
active_service = "remote"
[engine.service_destinations.remote]
uri = "ssh://core@build:22/run/podman/podman.sock"
identity = "/home/me/.ssh/id_ed25519"
`)
	activeSSH := writeConf(t, `[engine]
active_service = "remote"
[engine.service_destinations.remote]
uri = "ssh://core@build:22/run/podman/podman.sock"
[engine.service_destinations.local]
uri = "unix:///run/user/1000/podman/podman.sock"
`)
	runtimeDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(runtimeDir, "podman"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(runtimeDir, "podman", "podman.sock"), nil, 0o600))

	// podman 5 on macOS puts the machine's API socket under $TMPDIR
	machineSock := filepath.Join(t.TempDir(), "podman-machine-default-api.sock")
	require.NoError(t, os.WriteFile(machineSock, nil, 0o600))

	tests := []struct {
		name          string
		containerHost string
		configs       []string
		runtimeDir    string
		machineSocket string
		want          string
		wantErr       string
	}{
		{name: "CONTAINER_HOST unix", containerHost: "unix:///tmp/podman.sock", configs: []string{activeSSH}, want: "unix:///tmp/podman.sock"},
		{name: "CONTAINER_HOST ssh", containerHost: "ssh://core@build:22/run/podman/podman.sock", runtimeDir: runtimeDir, wantErr: "ssh"},
		{name: "config unix destination over an active ssh one", configs: []string{activeSSH}, want: "unix:///run/user/1000/podman/podman.sock"},
		{name: "config with only ssh", configs: []string{sshOnly}, runtimeDir: t.TempDir(), wantErr: "no podman socket"},
		{name: "rootless default socket", configs: []string{sshOnly}, runtimeDir: runtimeDir, machineSocket: machineSock, want: "unix://" + filepath.Join(runtimeDir, "podman", "podman.sock")},
		{name: "macOS machine with only ssh in config", configs: []string{sshOnly}, runtimeDir: t.TempDir(), machineSocket: machineSock, want: "unix://" + machineSock},
		{name: "machine socket missing (machine stopped)", configs: []string{sshOnly}, runtimeDir: t.TempDir(), machineSocket: filepath.Join(t.TempDir(), "gone.sock"), wantErr: "no podman socket"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CONTAINER_HOST", tt.containerHost)
			got, err := podmanHost(tt.configs, tt.runtimeDir, func() string { return tt.machineSocket })
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
