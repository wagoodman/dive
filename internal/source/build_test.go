package source

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestParseBuildArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantCtx     string
		wantHasFile bool
	}{
		{name: "context only", args: []string{"."}, wantCtx: "."},
		{name: "tag value is not the context", args: []string{"-t", "app", "."}, wantCtx: "."},
		{name: "context before flags", args: []string{".", "-t", "app"}, wantCtx: "."},
		{name: "bool flags take no value", args: []string{"--no-cache", "--pull", "-q", "ctx"}, wantCtx: "ctx"},
		{name: "inline values", args: []string{"--tag=app", "--build-arg=A=b", "ctx"}, wantCtx: "ctx"},
		{name: "stdin context", args: []string{"-t", "app", "-"}, wantCtx: "-"},
		{name: "url context", args: []string{"https://github.com/org/repo.git#main"}, wantCtx: "https://github.com/org/repo.git#main"},
		{name: "no context", args: []string{"-t", "app"}},
		{name: "-f", args: []string{"-f", "build/Dockerfile", "."}, wantCtx: ".", wantHasFile: true},
		{name: "--file", args: []string{"--file", "build/Dockerfile", "."}, wantCtx: ".", wantHasFile: true},
		{name: "--file=", args: []string{"--file=build/Dockerfile", "."}, wantCtx: ".", wantHasFile: true},
		{name: "-f attached", args: []string{"-fbuild/Dockerfile", "."}, wantCtx: ".", wantHasFile: true},
		{name: "-f=", args: []string{"-f=build/Dockerfile", "."}, wantCtx: ".", wantHasFile: true},
		{name: "-f stdin", args: []string{"-f", "-", "."}, wantCtx: ".", wantHasFile: true},
		{name: "similar long flag", args: []string{"--file-format=json", "."}, wantCtx: "."},
		{name: "-f as a flag value", args: []string{"--label", "-f", "."}, wantCtx: "."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, hasFile := parseBuildArgs(tt.args)
			assert.Equal(t, tt.wantCtx, ctx)
			assert.Equal(t, tt.wantHasFile, hasFile)
		})
	}
}

func TestFindContainerfile(t *testing.T) {
	tests := []struct {
		name     string
		buildCtx string
		files    []string
		want     string
	}{
		{name: "Containerfile", buildCtx: "ctx", files: []string{"ctx/Containerfile"}, want: filepath.Join("ctx", "Containerfile")},
		{name: "containerfile", buildCtx: "ctx", files: []string{"ctx/containerfile"}, want: filepath.Join("ctx", "containerfile")},
		{name: "Dockerfile", buildCtx: "ctx", files: []string{"ctx/Dockerfile"}, want: filepath.Join("ctx", "Dockerfile")},
		{name: "dockerfile", buildCtx: "ctx", files: []string{"ctx/dockerfile"}, want: filepath.Join("ctx", "dockerfile")},
		{name: "Containerfile over Dockerfile", buildCtx: "ctx", files: []string{"ctx/Containerfile", "ctx/Dockerfile"}, want: filepath.Join("ctx", "Containerfile")},
		{name: "only the context is searched", buildCtx: "ctx", files: []string{"app/Dockerfile", "ctx/main.go"}},
		{name: "missing context", buildCtx: "nope"},
		{name: "context is a file", buildCtx: "ctx.txt", files: []string{"ctx.txt"}},
		{name: "stdin", buildCtx: "-"},
		{name: "url", buildCtx: "https://github.com/org/repo.git"},
		{name: "no context"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			for _, f := range tt.files {
				create(t, fs, f, "FROM alpine")
			}
			assert.Equal(t, tt.want, findContainerfile(fs, tt.buildCtx))
		})
	}
}

func create(t testing.TB, fs afero.Fs, path, contents string) {
	t.Helper()

	dir := filepath.Dir(path)
	if dir != "." {
		err := fs.MkdirAll(dir, 0755)
		require.NoError(t, err, "Failed to create directory: %s", dir)
	}

	err := afero.WriteFile(fs, path, []byte(contents), 0644)
	require.NoError(t, err, "Failed to write file: %s", path)
}

// fakeEngine puts an executable named engine on PATH that writes id to the --iidfile it is given. It returns a func
// reporting the args the engine was last run with (one per line).
func fakeEngine(t *testing.T, engine, id string) func() []string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake engine is a shell script")
	}
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > '" + argsFile + "'\nwhile [ $# -gt 0 ]; do\n  if [ \"$1\" = \"--iidfile\" ]; then printf '" + id + "\\n' > \"$2\"; fi\n  shift\ndone\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, engine), []byte(script), 0o755))
	t.Setenv("PATH", dir)
	return func() []string {
		b, err := os.ReadFile(argsFile)
		require.NoError(t, err)
		return strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	}
}

func TestBuild(t *testing.T) {
	const hex = "9f5a0b6c1d2e3f40112233445566778899aabbccddeeff00112233445566778a"
	tests := []struct {
		engine  string
		args    []string
		wantRef string
		wantErr bool
	}{
		{engine: Docker, args: []string{"-f", "Dockerfile", "."}, wantRef: "docker://sha256:" + hex},
		{engine: Podman, args: []string{"."}, wantRef: "podman://" + hex},
		{engine: Containerd, args: []string{"."}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.engine+" "+tt.wantRef, func(t *testing.T) {
			fakeEngine(t, tt.engine, "sha256:"+hex)
			ref, err := Build(context.Background(), tt.engine, tt.args)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantRef, ref)

			// the ref must resolve back to the engine and the ID
			src, r := Parse(ref, Auto)
			assert.Equal(t, tt.engine, src)
			assert.Equal(t, strings.TrimPrefix(tt.wantRef, tt.engine+"://"), r)
		})
	}
}

func TestBuild_dockerfileDetection(t *testing.T) {
	withContainerfile := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(withContainerfile, "Containerfile"), []byte("FROM alpine"), 0o600))

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "Containerfile in the context", args: []string{"-t", "app", withContainerfile}, want: []string{"-f", filepath.Join(withContainerfile, "Containerfile"), "-t", "app", withContainerfile}},
		{name: "--file= given", args: []string{"--file=x/Dockerfile", withContainerfile}, want: []string{"--file=x/Dockerfile", withContainerfile}},
		// docker finds the Dockerfile itself, or reports it missing
		{name: "nothing in the context", args: []string{t.TempDir()}, want: nil},
		{name: "stdin", args: []string{"-"}, want: []string{"-"}},
		{name: "url", args: []string{"https://github.com/org/repo.git"}, want: []string{"https://github.com/org/repo.git"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := fakeEngine(t, Docker, "sha256:abc")
			_, err := Build(context.Background(), Docker, tt.args)
			require.NoError(t, err)
			want := tt.want
			if want == nil {
				want = tt.args
			}
			assert.Equal(t, want, args()[3:])
		})
	}
}

func TestBuild_passesArgsUntouched(t *testing.T) {
	args := fakeEngine(t, Podman, "sha256:abc")
	_, err := Build(context.Background(), Podman, []string{"--label", "", "--build-arg", "PAD= ", "-t", "x", "."})
	require.NoError(t, err)
	got := args()
	// "build --iidfile <file>" come first
	require.Greater(t, len(got), 3)
	assert.Equal(t, []string{"--label", "", "--build-arg", "PAD= ", "-t", "x", "."}, got[3:])
}
