package filetree

import (
	"testing"

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

func TestComparer_Marks(t *testing.T) {
	tests := []struct {
		fixture string
		key     TreeIndexKey
		want    map[string]DiffType
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
			cmp := NewComparer(imagefixture.Load(t, tt.fixture))
			tree, err := cmp.GetTree(tt.key)
			require.NoError(t, err)

			var paths []string
			for p := range tt.want {
				paths = append(paths, p)
			}
			assert.Equal(t, tt.want, diffTypes(t, tree, paths...))

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
