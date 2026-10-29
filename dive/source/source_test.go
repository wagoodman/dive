package source

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/anchore/stereoscope/pkg/file"
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

		// stereoscope ":" schemes
		{"docker:alpine", Podman, Docker, "alpine"},
		{"podman:alpine:3.20", Docker, Podman, "alpine:3.20"},
		{"docker-archive:img.tar", Docker, DockerArchive, "img.tar"},
		{"docker-tar:img.tar", Docker, DockerArchive, "img.tar"},
		{"oci-archive:img.tar", Docker, OCIArchive, "img.tar"},
		{"oci-dir:/abs/layout", Docker, OCIDir, "/abs/layout"},
		{"registry:alpine", Docker, Registry, "alpine"},
		{"oci-registry:alpine", Docker, Registry, "alpine"},
		{"containerd:alpine", Docker, Containerd, "alpine"},
		{"containers-storage:localhost/img:latest", Docker, ContainersStorage, "localhost/img:latest"},
		{"registry:localhost:5000/img:tag", Docker, Registry, "localhost:5000/img:tag"},

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
		{src: Podman, goos: "linux", want: []string{"containers-storage", "podman"}},
		{src: Podman, goos: "darwin", want: []string{"podman"}},
		{src: Podman, goos: "windows", want: []string{"podman"}},
		{src: DockerArchive, goos: "linux", want: []string{"docker-archive"}},
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
			got, err := dockerHost(newConfigDir(t, tt.currentContext, contexts))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDockerHost_badConfig(t *testing.T) {
	t.Setenv("DOCKER_HOST", "")
	t.Setenv("DOCKER_CONTEXT", "")
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte("{not json"), 0o600))
	_, err := dockerHost(dir)
	require.Error(t, err)
}
