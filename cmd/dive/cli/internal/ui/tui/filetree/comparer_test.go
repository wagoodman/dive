package filetree

import (
	"archive/tar"
	"testing"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/image"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/internal/imagefixture"
)

func diffTypes(t *testing.T, tree *FileTree, paths ...string) map[string]DiffType {
	t.Helper()
	got := make(map[string]DiffType)
	for _, p := range paths {
		n, err := tree.GetNode(p)
		require.NoError(t, err, p)
		got[p] = n.Data.DiffType
	}
	return got
}

// opaqueLayers makes /etc/app opaque in layer 2 without adding anything there, after layer 1 added b.conf, then
// re-adds a.conf in layer 3
var opaqueLayers = imagefixture.Spec{Layers: [][]imagefixture.Entry{
	{
		{Name: "etc/", Type: tar.TypeDir},
		{Name: "etc/app/", Type: tar.TypeDir},
		{Name: "etc/app/a.conf", Content: "a"},
		{Name: "etc/app/sub/", Type: tar.TypeDir},
		{Name: "etc/app/sub/c.conf", Content: "c"},
		{Name: "etc/keep.conf", Content: "keep"},
	},
	{
		{Name: "etc/", Type: tar.TypeDir},
		{Name: "etc/app/", Type: tar.TypeDir},
		{Name: "etc/app/b.conf", Content: "b"},
	},
	{
		{Name: "etc/", Type: tar.TypeDir},
		{Name: "etc/app/", Type: tar.TypeDir},
		{Name: "etc/app/.wh..wh..opq"},
	},
	{
		{Name: "etc/", Type: tar.TypeDir},
		{Name: "etc/app/", Type: tar.TypeDir},
		{Name: "etc/app/a.conf", Content: "a"},
	},
}}

// linkRetarget points a symlink and a hardlink somewhere else, with the same type, mode, owner and size, so only
// the link name differs
var linkRetarget = imagefixture.Spec{Layers: [][]imagefixture.Entry{
	{
		{Name: "bin/", Type: tar.TypeDir},
		{Name: "bin/a", Content: "a"},
		{Name: "bin/b", Content: "b"},
		{Name: "bin/sym", Type: tar.TypeSymlink, Linkname: "a"},
		{Name: "bin/same", Type: tar.TypeSymlink, Linkname: "a"},
		{Name: "bin/hard", Type: tar.TypeLink, Linkname: "bin/a"},
	},
	{
		{Name: "bin/", Type: tar.TypeDir},
		{Name: "bin/sym", Type: tar.TypeSymlink, Linkname: "b"},
		{Name: "bin/same", Type: tar.TypeSymlink, Linkname: "a"},
		{Name: "bin/hard", Type: tar.TypeLink, Linkname: "bin/b"},
	},
}}

// implicitReAdd is implicit-parent-dirs plus a layer re-adding the deep file unchanged
var implicitReAdd = func() imagefixture.Spec {
	spec := imagefixture.Fixtures()["implicit-parent-dirs"]
	spec.Layers = append(spec.Layers, []imagefixture.Entry{{Name: "a/b/c/deep.txt", Content: "deep\n"}})
	return spec
}()

func TestComparer_Marks(t *testing.T) {
	tests := []struct {
		fixture string
		// spec is loaded instead of the named fixture when set
		spec *imagefixture.Spec
		key  TreeIndexKey
		want map[string]DiffType
		// absent paths are not in the tree at all
		absent []string
	}{
		{
			fixture: "legacy-save",
			key:     NewTreeIndexKey(0, 1, 2, 2),
			want: map[string]DiffType{
				// content only change, found by hashing
				"/app/main.txt":    Modified,
				"/app/config.yaml": Unmodified,
				"/etc/os-release":  Removed,
				"/bin/sh":          Unmodified,
			},
		},
		{
			fixture: "legacy-save",
			key:     NewTreeIndexKey(0, 0, 1, 1),
			want: map[string]DiffType{
				"/app":          Added,
				"/app/main.txt": Added,
				"/bin/sh":       Unmodified,
			},
		},
		{
			fixture: "mode-uid-only-change",
			key:     NewTreeIndexKey(0, 0, 1, 1),
			want: map[string]DiffType{
				"/app/run.sh": Modified,
				"/app/owned":  Modified,
				"/app/same":   Unmodified,
				"/app":        Modified,
			},
		},
		{
			fixture: "opaque-whiteout-dir",
			// the opaque dir hides everything layer 0 had in it
			key: NewTreeIndexKey(0, 0, 1, 1),
			want: map[string]DiffType{
				"/etc/app/a.conf":     Removed,
				"/etc/app/b.conf":     Removed,
				"/etc/app/sub":        Removed,
				"/etc/app/sub/c.conf": Removed,
				"/etc/app/new.conf":   Added,
				"/etc/app":            Modified,
				"/etc/keep.conf":      Unmodified,
			},
		},
		{
			fixture: "opaque-layers",
			spec:    &opaqueLayers,
			// the opaque layer itself adds nothing, the ancestors still change as for a plain whiteout
			key: NewTreeIndexKey(0, 1, 2, 2),
			want: map[string]DiffType{
				"/etc/app/a.conf":     Removed,
				"/etc/app/b.conf":     Removed,
				"/etc/app/sub":        Removed,
				"/etc/app/sub/c.conf": Removed,
				"/etc/app":            Modified,
				"/etc":                Modified,
				"/etc/keep.conf":      Unmodified,
			},
		},
		{
			fixture: "opaque-layers",
			spec:    &opaqueLayers,
			// the layer after: what the opaque dir hid is gone from the bottom tree, not Removed again
			key: NewTreeIndexKey(0, 2, 3, 3),
			want: map[string]DiffType{
				"/etc/app/a.conf": Added,
				"/etc/app":        Modified,
				"/etc/keep.conf":  Unmodified,
			},
			absent: []string{"/etc/app/b.conf", "/etc/app/sub"},
		},
		{
			fixture: "opaque-layers",
			spec:    &opaqueLayers,
			// compare all: b.conf added by layer 1 is hidden by layer 2
			key: NewTreeIndexKey(0, 0, 1, 2),
			want: map[string]DiffType{
				"/etc/app/a.conf":     Removed,
				"/etc/app/b.conf":     Removed,
				"/etc/app/sub/c.conf": Removed,
				"/etc/app":            Modified,
				"/etc":                Modified,
				"/etc/keep.conf":      Unmodified,
			},
		},
		{
			fixture: "link-retarget",
			spec:    &linkRetarget,
			key:     NewTreeIndexKey(0, 0, 1, 1),
			want: map[string]DiffType{
				"/bin/sym":  Modified,
				"/bin/hard": Modified,
				"/bin/same": Unmodified,
			},
		},
		{
			fixture: "implicit-parent-dirs re-add",
			spec:    &implicitReAdd,
			// the implicit dirs on both sides have no catalog entry, they compare equal
			key: NewTreeIndexKey(0, 1, 2, 2),
			want: map[string]DiffType{
				"/a":              Unmodified,
				"/a/b":            Unmodified,
				"/a/b/c":          Unmodified,
				"/a/b/c/deep.txt": Unmodified,
				"/a/b/other.txt":  Unmodified,
			},
		},
		{
			fixture: "hardlink-and-symlink",
			key:     NewTreeIndexKey(0, 0, 1, 1),
			want: map[string]DiffType{
				// a symlink replaced by a file
				"/usr/bin/tool-rel":  Modified,
				"/usr/bin/libx-hard": Added,
				"/usr/bin/tool":      Unmodified,
				"/usr/bin/tool-hard": Unmodified,
				"/lib":               Unmodified,
			},
		},
		{
			fixture: "file-and-dir-whiteouts",
			// compare all: layers 1 and 2 against layer 0
			key: NewTreeIndexKey(0, 0, 1, 2),
			want: map[string]DiffType{
				"/data/remove.txt":      Removed,
				"/data/olddir/x":        Removed,
				"/data/olddir/nested/y": Removed,
				"/data/olddir/z":        Added,
				"/data/keep.txt":        Unmodified,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.fixture+" "+tt.key.String(), func(t *testing.T) {
			var img *image.Image
			if tt.spec != nil {
				img = imagefixture.LoadSpec(t, *tt.spec)
			} else {
				img = imagefixture.Load(t, tt.fixture)
			}
			cmp := NewComparer(img)
			tree, err := cmp.GetTree(tt.key)
			require.NoError(t, err)

			var paths []string
			for p := range tt.want {
				paths = append(paths, p)
			}
			assert.Equal(t, tt.want, diffTypes(t, tree, paths...))
			for _, p := range tt.absent {
				_, err := tree.GetNode(p)
				assert.Error(t, err, p)
			}

			// trees are built once per key
			again, err := cmp.GetTree(tt.key)
			require.NoError(t, err)
			assert.Same(t, tree, again)
		})
	}
}

func TestComparer_HashesOnlyWhenNeeded(t *testing.T) {
	cmp := NewComparer(imagefixture.Load(t, "mode-uid-only-change"))
	_, err := cmp.GetTree(NewTreeIndexKey(0, 0, 1, 1))
	require.NoError(t, err)
	// run.sh and owned differ by metadata, so only the two copies of /app/same get read
	assert.Len(t, cmp.hashes, 2)
}

func TestComparer_HashFailureIsModified(t *testing.T) {
	cmp := NewComparer(imagefixture.Load(t, "mode-uid-only-change"))
	// a reference the catalog doesn't know can't be read
	missing := file.NewFileReference("/nope")
	f := FileInfo{TypeFlag: tar.TypeReg, Size: 4, ref: missing}

	_, ok := cmp.contentHash(f)
	assert.False(t, ok)
	assert.Empty(t, cmp.hashes, "a failed read is retried, not cached")
	assert.Equal(t, Modified, f.Compare(f, cmp.contentHash))
}

func TestFileInfo_Compare(t *testing.T) {
	noHash := func(t *testing.T) ContentHasher {
		return func(FileInfo) (uint64, bool) {
			t.Error("content should not be read")
			return 0, false
		}
	}
	reg := FileInfo{TypeFlag: tar.TypeReg, Size: 10}

	bigger := reg
	bigger.Size = 11
	assert.Equal(t, Modified, reg.Compare(bigger, noHash(t)), "a size change needs no hashing")

	link := FileInfo{TypeFlag: tar.TypeSymlink, Linkname: "a"}
	retargeted := link
	retargeted.Linkname = "b"
	assert.Equal(t, Modified, link.Compare(retargeted, noHash(t)))
	assert.Equal(t, Unmodified, link.Compare(link, noHash(t)), "links have no content to read")

	same := func(FileInfo) (uint64, bool) { return 1, true }
	assert.Equal(t, Unmodified, reg.Compare(reg, same))
	assert.Equal(t, Modified, reg.Compare(reg, func(FileInfo) (uint64, bool) { return 0, false }))
}
