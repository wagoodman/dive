package app

import (
	"archive/tar"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/anchore/stereoscope"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/internal/imagefixture"
)

func loadSpec(t *testing.T, spec imagefixture.Spec) *image.Image {
	t.Helper()
	p := filepath.Join(t.TempDir(), "image.tar")
	f, err := os.Create(p)
	require.NoError(t, err)
	require.NoError(t, spec.Write(f))
	require.NoError(t, f.Close())

	img, err := stereoscope.GetImageFromSource(context.Background(), p, image.DockerTarballSource)
	require.NoError(t, err)
	t.Cleanup(func() { _ = img.Cleanup() })
	return img
}

// dirs runs the test with a fresh cwd (base/dst) next to a dir that must stay untouched (base/outside)
func dirs(t *testing.T) (base, outside string) {
	base = t.TempDir()
	outside = filepath.Join(base, "outside")
	require.NoError(t, os.Mkdir(outside, 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(base, "dst"), 0o755))
	t.Chdir(filepath.Join(base, "dst"))
	return base, outside
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	return string(b)
}

func TestExtract(t *testing.T) {
	img := imagefixture.Load(t, "test-docker-image")
	dirs(t)

	// layer 4 adds root/example/somefile1.txt (the chmod one)
	require.NoError(t, extract(img, 4, "/root", "."))

	got, err := os.ReadFile(filepath.Join("root", "example", "somefile1.txt"))
	require.NoError(t, err)
	assert.Len(t, got, 6405)

	// only the selected layer, only under the selected path
	_, err = os.Stat(filepath.Join("root", "saved.txt"))
	assert.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat("somefile.txt")
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestExtract_SelectsPathNotPrefix(t *testing.T) {
	img := loadSpec(t, imagefixture.Spec{Layers: [][]imagefixture.Entry{{
		{Name: "app/", Type: tar.TypeDir},
		{Name: "app/a.txt", Content: "a"},
		{Name: "application/", Type: tar.TypeDir},
		{Name: "application/b.txt", Content: "b"},
		{Name: "app/link", Type: tar.TypeSymlink, Linkname: "/etc/passwd"},
		{Name: "app/hard", Type: tar.TypeLink, Linkname: "app/a.txt"},
	}}})
	dirs(t)

	require.NoError(t, extract(img, 0, "/app", "."))

	assert.Equal(t, "a", readFile(t, filepath.Join("app", "a.txt")))
	// a string prefix match would have picked up /application
	_, err := os.Stat("application")
	assert.ErrorIs(t, err, os.ErrNotExist)
	// links are not written, as before
	_, err = os.Lstat(filepath.Join("app", "link"))
	assert.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Lstat(filepath.Join("app", "hard"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestExtract_DoesNotEscape(t *testing.T) {
	img := loadSpec(t, imagefixture.Spec{Layers: [][]imagefixture.Entry{{
		{Name: "../escape.txt", Content: "escape"},
		{Name: "/abs.txt", Content: "abs"},
		{Name: "ok/../../up.txt", Content: "up"},
		// a symlinked dir in the layer, then a file through it
		{Name: "evil", Type: tar.TypeSymlink, Linkname: "../outside"},
		{Name: "evil/pwned.txt", Content: "pwned"},
		// cwd already has symlinks here, pointing out of it
		{Name: "linked/", Type: tar.TypeDir},
		{Name: "linked/file.txt", Content: "linked"},
		{Name: "target.txt", Content: "target"},
		{Name: "fine.txt", Content: "fine"},
	}}})
	base, outside := dirs(t)
	require.NoError(t, os.Symlink(outside, "linked"))
	require.NoError(t, os.Symlink(filepath.Join(outside, "target.txt"), "target.txt"))

	err := extract(img, 0, "/", ".")
	require.Error(t, err)
	for _, refused := range []string{"../escape.txt", "/abs.txt", "ok/../../up.txt", "/linked/file.txt", "/target.txt"} {
		assert.ErrorContains(t, err, refused)
	}

	// nothing outside of the destination
	entries, err := os.ReadDir(outside)
	require.NoError(t, err)
	assert.Empty(t, entries, "files were written outside the destination")
	for _, name := range []string{"escape.txt", "abs.txt", "up.txt", "pwned.txt"} {
		_, err := os.Stat(filepath.Join(base, name))
		assert.ErrorIs(t, err, os.ErrNotExist, name)
	}
	_, err = os.Stat("/abs.txt")
	assert.ErrorIs(t, err, os.ErrNotExist)

	// entries that are safe still land
	assert.Equal(t, "fine", readFile(t, "fine.txt"))
}
