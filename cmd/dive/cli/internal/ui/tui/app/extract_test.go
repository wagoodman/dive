package app

import (
	"archive/tar"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/internal/imagefixture"
)

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

	// the tree as of the selected layer (root/saved.txt is added by a later one), only under the selected path
	_, err = os.Stat(filepath.Join("root", "saved.txt"))
	assert.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat("somefile.txt")
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestExtract_SelectsPathNotPrefix(t *testing.T) {
	img := imagefixture.LoadSpec(t, imagefixture.Spec{Layers: [][]imagefixture.Entry{{
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
	img := imagefixture.LoadSpec(t, imagefixture.Spec{Layers: [][]imagefixture.Entry{{
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

func TestExtract_SquashedTree(t *testing.T) {
	img := imagefixture.LoadSpec(t, imagefixture.Spec{Layers: [][]imagefixture.Entry{
		{
			{Name: "app/", Type: tar.TypeDir},
			{Name: "app/a.txt", Content: "a"},
			{Name: "app/gone.txt", Content: "gone"},
		},
		{
			{Name: "app/", Type: tar.TypeDir},
			{Name: "app/b.txt", Content: "b"},
			{Name: "app/.wh.gone.txt"},
		},
	}})
	dirs(t)

	// layer 1 shows a.txt from layer 0 too, and not what it removed
	require.NoError(t, extract(img, 1, "/app", "."))
	assert.Equal(t, "a", readFile(t, filepath.Join("app", "a.txt")))
	assert.Equal(t, "b", readFile(t, filepath.Join("app", "b.txt")))
	for _, name := range []string{"gone.txt", ".wh.gone.txt"} {
		_, err := os.Lstat(filepath.Join("app", name))
		assert.ErrorIs(t, err, os.ErrNotExist, name)
	}

	// nothing there is reported, not silently ignored
	assert.ErrorContains(t, extract(img, 1, "/nope", "."), "no regular files to extract")
}

func TestExtract_KeepsExistingFiles(t *testing.T) {
	img := imagefixture.LoadSpec(t, imagefixture.Spec{Layers: [][]imagefixture.Entry{{
		{Name: "config.yaml", Content: "from the image"},
		{Name: "new.txt", Content: "new"},
	}}})
	dirs(t)
	// someone else's read-only file, which only dir write permission would let us unlink
	require.NoError(t, os.WriteFile("config.yaml", []byte("mine"), 0o444))

	err := extract(img, 0, "/", ".")
	assert.ErrorContains(t, err, `"config.yaml" exists, not overwritten`)

	assert.Equal(t, "mine", readFile(t, "config.yaml"))
	fi, err := os.Stat("config.yaml")
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o444), fi.Mode().Perm())
	assert.Equal(t, "new", readFile(t, "new.txt"))
}
