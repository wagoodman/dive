package cli

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/cmd/dive/cli/internal/command/export"
	"github.com/wagoodman/dive/internal/imagefixture"
)

// these run the real command against image archives, so the CLI path is covered without a container engine

var archiveFixtures = []struct {
	fixture, scheme string
}{
	{fixture: "test-docker-image", scheme: "docker-archive:"},
	{fixture: "test-oci-docker-image", scheme: "docker-archive://"},
	{fixture: "oci-archive-gzip", scheme: "oci-archive:"},
	{fixture: "opaque-whiteout-dir", scheme: "docker-archive:"},
}

func Test_Archive_JsonOutput(t *testing.T) {
	for _, tt := range archiveFixtures {
		t.Run(tt.fixture, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "output.json")
			// a longer file at the destination must not leave trailing bytes behind
			require.NoError(t, os.WriteFile(file, make([]byte, 1<<20), 0o644))

			rootCmd := getTestCommand(t, tt.scheme+imagefixture.Path(t, tt.fixture)+" --json "+file)
			combined := Capture().WithStdout().WithStderr().Run(t, func() {
				require.NoError(t, rootCmd.Execute())
			})
			assert.Contains(t, combined, "Exporting details")

			got, err := os.ReadFile(file)
			require.NoError(t, err)
			want, err := os.ReadFile(imagefixture.GoldenPath(tt.fixture, "export"))
			require.NoError(t, err)
			assert.JSONEq(t, string(want), string(got))
		})
	}
}

func Test_Archive_CI(t *testing.T) {
	archive := "docker-archive:" + repoPath(t, ".data/test-docker-image.tar")

	t.Run("fail", func(t *testing.T) {
		t.Setenv("DIVE_CONFIG", "./testdata/image-multi-layer-dockerfile/dive-fail.yaml")
		rootCmd := getTestCommand(t, archive)
		stdout := Capture().WithStdout().WithSuppress().Run(t, func() {
			require.ErrorContains(t, rootCmd.Execute(), "evaluation failed")
		})
		snaps.MatchSnapshot(t, stdout)
	})

	t.Run("oci-archive", func(t *testing.T) {
		rootCmd := getTestCommand(t, "oci-archive:"+imagefixture.Path(t, "oci-archive-gzip"))
		stdout := Capture().WithStdout().WithSuppress().Run(t, func() {
			require.NoError(t, rootCmd.Execute())
		})
		snaps.MatchSnapshot(t, stdout)
	})

	t.Run("load failure", func(t *testing.T) {
		rootCmd := getTestCommand(t, "docker-archive:"+filepath.Join(t.TempDir(), "missing.tar"))
		combined := Capture().WithStdout().WithStderr().Run(t, func() {
			require.ErrorContains(t, rootCmd.Execute(), "cannot load image")
		})
		assert.Contains(t, combined, "Loading image")
	})
}

// containerdStoreSave writes spec (an OCISave) the way `docker save` lays out an image pulled into the containerd
// image store: index.json points at the image's multi-platform index instead of a single manifest, the other
// platform's manifest has no blobs in the archive, and buildkit's attestation manifest (unknown/unknown platform,
// in-toto layer) sits next to the image. The host platform is the one that was saved, as docker does.
func containerdStoreSave(t *testing.T, spec imagefixture.Spec) string {
	t.Helper()
	require.Equal(t, imagefixture.OCISave, spec.Format)

	var buf bytes.Buffer
	require.NoError(t, spec.Write(&buf))
	files := map[string][]byte{}
	dirs := map[string]bool{}
	tr := tar.NewReader(&buf)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if hdr.Typeflag == tar.TypeDir {
			dirs[hdr.Name] = true
			continue
		}
		files[hdr.Name], err = io.ReadAll(tr)
		require.NoError(t, err)
	}

	type platform struct {
		Architecture string `json:"architecture"`
		OS           string `json:"os"`
	}
	type descriptor struct {
		MediaType   string            `json:"mediaType"`
		Digest      string            `json:"digest"`
		Size        int               `json:"size"`
		Platform    *platform         `json:"platform,omitempty"`
		Annotations map[string]string `json:"annotations,omitempty"`
	}
	addBlob := func(v any) descriptor {
		data, ok := v.([]byte)
		if !ok {
			var err error
			data, err = json.Marshal(v)
			require.NoError(t, err)
		}
		sum := sha256.Sum256(data)
		files["blobs/sha256/"+hex.EncodeToString(sum[:])] = data
		return descriptor{Digest: "sha256:" + hex.EncodeToString(sum[:]), Size: len(data)}
	}

	var index struct {
		Manifests []descriptor `json:"manifests"`
	}
	require.NoError(t, json.Unmarshal(files["index.json"], &index))
	require.Len(t, index.Manifests, 1)
	image := index.Manifests[0]
	image.Annotations = nil
	image.Platform = &platform{Architecture: runtime.GOARCH, OS: "linux"}

	other := "arm64"
	if runtime.GOARCH == "arm64" {
		other = "amd64"
	}
	missing := descriptor{
		MediaType: "application/vnd.oci.image.manifest.v1+json",
		Digest:    "sha256:" + strings.Repeat("0", 64),
		Size:      1234,
		Platform:  &platform{Architecture: other, OS: "linux"},
	}

	statement := addBlob(map[string]any{"_type": "https://in-toto.io/Statement/v0.1", "predicateType": "https://slsa.dev/provenance/v0.2", "subject": []any{}, "predicate": map[string]any{}})
	statement.MediaType = "application/vnd.in-toto+json"
	statement.Annotations = map[string]string{"in-toto.io/predicate-type": "https://slsa.dev/provenance/v0.2"}
	attConfig := addBlob(map[string]any{"architecture": "unknown", "os": "unknown", "config": map[string]any{}, "rootfs": map[string]any{"type": "layers", "diff_ids": []string{statement.Digest}}})
	attConfig.MediaType = "application/vnd.oci.image.config.v1+json"
	attestation := addBlob(map[string]any{"schemaVersion": 2, "mediaType": "application/vnd.oci.image.manifest.v1+json", "config": attConfig, "layers": []descriptor{statement}})
	attestation.MediaType = "application/vnd.oci.image.manifest.v1+json"
	attestation.Platform = &platform{Architecture: "unknown", OS: "unknown"}
	attestation.Annotations = map[string]string{
		"vnd.docker.reference.digest": image.Digest,
		"vnd.docker.reference.type":   "attestation-manifest",
	}

	imageIndex := addBlob(map[string]any{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.index.v1+json",
		"manifests":     []descriptor{image, missing, attestation},
	})
	imageIndex.MediaType = "application/vnd.oci.image.index.v1+json"
	imageIndex.Annotations = map[string]string{
		"io.containerd.image.name":          "docker.io/library/test:latest",
		"org.opencontainers.image.ref.name": "latest",
	}
	var err error
	files["index.json"], err = json.Marshal(map[string]any{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.index.v1+json",
		"manifests":     []descriptor{imageIndex},
	})
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "containerd-store-save.tar")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	tw := tar.NewWriter(f)
	for _, name := range slices.Sorted(maps.Keys(dirs)) {
		require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeDir, Mode: 0o755}))
	}
	for _, name := range slices.Sorted(maps.Keys(files)) {
		data := files[name]
		require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(data))}))
		_, err := tw.Write(data)
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	return path
}

// regression for #718 and #452: `docker save` output from docker 25+ and from the containerd image store failed to
// load ("could not find image config", "could not find 'blobs/sha256/...' in parsed layers"). The save here has
// gzip layers (one a tiny near empty layer), the image's multi-platform index with a missing platform, and an
// attestation manifest. It is read as a docker-archive and through a docker daemon, as `dive test` does.
func Test_Archive_ContainerdStoreSave(t *testing.T) {
	spec := imagefixture.Fixtures()["oci-save-gzip"]
	spec.Layers = append(spec.Layers, []imagefixture.Entry{{Name: "tmp/", Type: tar.TypeDir}})
	spec.History = append(spec.History, imagefixture.History{CreatedBy: "RUN mkdir -p /tmp # buildkit"})

	// the same image as a plain single manifest save is the reference
	want := exportArchive(t, "docker-archive:"+writeSpec(t, spec))
	require.Len(t, want.Layer, 4)
	assert.Equal(t, "RUN mkdir -p /tmp # buildkit", want.Layer[3].Command)
	require.Len(t, want.Layer[3].FileList, 1)
	assert.Equal(t, "tmp", want.Layer[3].FileList[0].Path)

	saved := containerdStoreSave(t, spec)

	t.Run("docker-archive", func(t *testing.T) {
		assert.Equal(t, want, exportArchive(t, "docker-archive:"+saved))
	})

	t.Run("docker daemon", func(t *testing.T) {
		saves := fakeDockerDaemon(t, saved)
		assert.Equal(t, want, exportArchive(t, "docker:test"))
		assert.Equal(t, int32(1), saves.Load())
	})
}

// exportArchive runs dive with --json on the reference and returns the export.
func exportArchive(t *testing.T, ref string) export.Export {
	t.Helper()
	file := filepath.Join(t.TempDir(), "output.json")
	rootCmd := getTestCommand(t, ref+" --json "+file)
	Capture().WithSuppress().Run(t, func() {
		require.NoError(t, rootCmd.Execute())
	})
	data, err := os.ReadFile(file)
	require.NoError(t, err)
	var got export.Export
	require.NoError(t, json.Unmarshal(data, &got))
	return got
}

// fakeDockerDaemon serves the docker API calls dive makes for an image that is already local (ping, inspect, save)
// from DOCKER_HOST, with the save returning the given tarball. Any other call (e.g. a pull) fails the test. It
// returns the number of saves served.
func fakeDockerDaemon(t *testing.T, tarball string) *atomic.Int32 {
	t.Helper()
	var saves atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Api-Version", "1.47")
		w.Header().Set("Ostype", "linux")
		switch {
		case strings.HasSuffix(r.URL.Path, "/_ping"):
			_, _ = io.WriteString(w, "OK")
		case strings.Contains(r.URL.Path, "/images/") && strings.HasSuffix(r.URL.Path, "/json"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"Id":"sha256:`+strings.Repeat("a", 64)+`","RepoTags":["test:latest"],"Os":"linux","Architecture":"`+runtime.GOARCH+`","Size":1}`)
		case strings.HasSuffix(r.URL.Path, "/images/get"):
			saves.Add(1)
			http.ServeFile(w, r, tarball)
		default:
			t.Errorf("unexpected docker API call: %s %s", r.Method, r.URL)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	t.Setenv("DOCKER_HOST", "tcp://"+srv.Listener.Addr().String())
	t.Setenv("DOCKER_CONTEXT", "")
	t.Setenv("DOCKER_TLS_VERIFY", "")
	t.Setenv("DOCKER_TLS", "")
	return &saves
}

// regression for #318: kaniko's `--tarPath` archive (written by go-containerregistry's tarball writer) failed with
// "could not find image config", since its config is named `sha256:<hex>` and its layers `<hex>.tar.gz`.
func Test_Archive_KanikoTarPath(t *testing.T) {
	plain := writeSpec(t, imagefixture.Fixtures()["legacy-save"])
	want := exportArchive(t, "docker-archive:"+plain)

	img, err := tarball.ImageFromPath(plain, nil)
	require.NoError(t, err)
	ref, err := name.ParseReference("image")
	require.NoError(t, err)
	kaniko := filepath.Join(t.TempDir(), "image.tar")
	require.NoError(t, tarball.MultiRefWriteToFile(kaniko, map[name.Reference]v1.Image{ref: img}))

	// the archive has kaniko's naming, not docker's
	manifest, err := tarball.LoadManifest(func() (io.ReadCloser, error) { return os.Open(kaniko) })
	require.NoError(t, err)
	require.Len(t, manifest, 1)
	assert.True(t, strings.HasPrefix(manifest[0].Config, "sha256:"), manifest[0].Config)
	for _, l := range manifest[0].Layers {
		assert.True(t, strings.HasSuffix(l, ".tar.gz"), l)
	}

	assert.Equal(t, want, exportArchive(t, "docker-archive://"+kaniko))
}

// writeSpec writes the spec's image tarball to a temp file.
func writeSpec(t *testing.T, spec imagefixture.Spec) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "image.tar")
	f, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, spec.Write(f))
	require.NoError(t, f.Close())
	return path
}
