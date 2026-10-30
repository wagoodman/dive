package layertree

import (
	"archive/tar"
	"io/fs"
	"testing"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/filetree/filenode"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/internal/imagefixture"
)

func TestWalk(t *testing.T) {
	img := imagefixture.Load(t, "opaque-whiteout-dir")
	walk := func(skip func(*filenode.FileNode) bool) (paths []string, leaves map[string]bool) {
		leaves = make(map[string]bool)
		require.NoError(t, Walk(img.Layers[1].Tree, skip, func(n *filenode.FileNode, leaf bool) error {
			paths = append(paths, string(n.RealPath))
			leaves[string(n.RealPath)] = leaf
			return nil
		}))
		return paths, leaves
	}

	// children before parents, in path order, the root left out
	paths, _ := walk(nil)
	assert.Equal(t, []string{"/etc/app/.wh..wh..opq", "/etc/app/new.conf", "/etc/app", "/etc"}, paths)

	// a skipped node is not a child either
	paths, leaves := walk(OpaqueMarker)
	assert.Equal(t, []string{"/etc/app/new.conf", "/etc/app", "/etc"}, paths)
	assert.True(t, leaves["/etc/app/new.conf"])
	assert.False(t, leaves["/etc/app"])
}

func TestHeader(t *testing.T) {
	img := imagefixture.Load(t, "hardlink-and-symlink")
	refs := make(map[string]*filenode.FileNode)
	require.NoError(t, Walk(img.Layers[0].Tree, nil, func(n *filenode.FileNode, _ bool) error {
		refs[string(n.RealPath)] = n
		return nil
	}))

	// stereoscope describes an in-layer hardlink as its target, the header still says hardlink
	hdr, ok := Header(img, refs["/usr/bin/tool-hard"].Reference)
	require.True(t, ok)
	assert.Equal(t, byte('1'), hdr.Typeflag)
	assert.Zero(t, hdr.Size)

	hdr, ok = Header(img, refs["/usr/bin/tool"].Reference)
	require.True(t, ok)
	assert.Equal(t, int64(1000), hdr.Size)

	_, ok = Header(img, nil)
	assert.False(t, ok)
}

func TestHeader_notTar(t *testing.T) {
	cat := image.NewFileCatalog()
	img := &image.Image{FileCatalog: cat}
	add := func(p string, ty file.Type, mode fs.FileMode, size int64, link string) *file.Reference {
		ref := file.NewFileReference(file.Path(p))
		cat.Add(*ref, file.Metadata{
			// no tar header in Sys, as for a squashfs source
			FileInfo:        file.ManualInfo{NameValue: p, SizeValue: size, ModeValue: mode},
			Path:            p,
			LinkDestination: link,
			UserID:          -1,
			GroupID:         -1,
			Type:            ty,
		}, nil, nil)
		return ref
	}

	tests := []struct {
		name string
		ref  *file.Reference
		want tar.Header
	}{
		{
			name: "regular",
			ref:  add("/bin/tool", file.TypeRegular, 0o755, 1000, ""),
			want: tar.Header{Typeflag: tar.TypeReg, Name: "bin/tool", Mode: 0o755, Size: 1000},
		},
		{
			// not the file it names, and 0 bytes
			name: "hardlink",
			ref:  add("/bin/tool-hard", file.TypeHardLink, 0o755, 1000, "/bin/tool"),
			want: tar.Header{Typeflag: tar.TypeLink, Name: "bin/tool-hard", Linkname: "/bin/tool", Mode: 0o755},
		},
		{
			name: "symlink",
			ref:  add("/bin/sh", file.TypeSymLink, fs.ModeSymlink|0o777, 4, "tool"),
			want: tar.Header{Typeflag: tar.TypeSymlink, Name: "bin/sh", Linkname: "tool", Mode: 0o777},
		},
		{
			// tar.FileInfoHeader refuses sockets
			name: "socket",
			ref:  add("/run/sock", file.TypeSocket, fs.ModeSocket|0o600, 0, ""),
			want: tar.Header{Typeflag: tar.TypeReg, Name: "run/sock", Mode: 0o600},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hdr, ok := Header(img, tt.ref)
			require.True(t, ok)
			assert.Equal(t, &tt.want, hdr)
		})
	}
}
