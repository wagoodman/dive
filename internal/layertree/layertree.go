// Package layertree has the stereoscope tree walking and file metadata lookups shared by analysis, export and the UI.
package layertree

import (
	"archive/tar"
	"sort"
	"strings"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/filetree"
	"github.com/anchore/stereoscope/pkg/filetree/filenode"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/anchore/stereoscope/pkg/tree/node"
)

// Walk visits every node of a layer tree depth first, children (in path order) before their parent, skipping the
// root. Nodes that skip reports true for are not visited and don't count as children, so they never make their
// parent a non-leaf. skip may be nil.
func Walk(tr filetree.Reader, skip func(*filenode.FileNode) bool, fn func(n *filenode.FileNode, leaf bool) error) error {
	t := tr.TreeReader()
	var visit func(n node.Node, isRoot bool) error
	visit = func(n node.Node, isRoot bool) error {
		var children node.Nodes
		for _, child := range t.Children(n) {
			if skip != nil && skip(child.(*filenode.FileNode)) {
				continue
			}
			children = append(children, child)
		}
		sort.Sort(children)
		for _, child := range children {
			if err := visit(child, false); err != nil {
				return err
			}
		}
		if isRoot {
			return nil
		}
		return fn(n.(*filenode.FileNode), len(children) == 0)
	}
	for _, root := range t.Roots() {
		if err := visit(root, true); err != nil {
			return err
		}
	}
	return nil
}

// OpaqueMarker reports whether n is an opaque whiteout marker (".wh..wh..opq"), which hides everything lower layers
// have in its directory.
func OpaqueMarker(n *filenode.FileNode) bool {
	return n.RealPath.IsDirWhiteout()
}

// Header returns the tar header the file at ref was read from. Stereoscope describes an in-layer hardlink as the file
// it names, the header still says hardlink. A source that isn't tar based gets a header built from the catalog
// metadata, so callers see one shape either way. It reports false when the catalog has no entry for ref.
func Header(img *image.Image, ref *file.Reference) (*tar.Header, bool) {
	if ref == nil {
		return nil, false
	}
	entry, err := img.FileCatalog.Get(*ref)
	if err != nil || entry.FileInfo == nil {
		return nil, false
	}
	if hdr, ok := entry.Sys().(*tar.Header); ok {
		return hdr, true
	}
	// ponytail: no dive source reaches this today (all are tar based), so permission bits only, no setuid/sticky
	hdr := &tar.Header{
		Typeflag: tarType(entry.Type),
		Name:     strings.TrimPrefix(entry.Path, "/"),
		Linkname: entry.LinkDestination,
		Mode:     int64(entry.Mode().Perm()),
		// squashfs has no owner in stereoscope's metadata (-1)
		Uid:     max(entry.UserID, 0),
		Gid:     max(entry.GroupID, 0),
		ModTime: entry.ModTime(),
	}
	// only regular files carry data, a hardlink is 0 bytes as in a tar
	if hdr.Typeflag == tar.TypeReg {
		hdr.Size = entry.Size()
	}
	return hdr, true
}

// tarType maps a stereoscope file type to its tar typeflag. Tar has no socket type, so a socket (or anything
// irregular) shows as an empty regular file rather than dropping out.
func tarType(t file.Type) byte {
	switch t {
	case file.TypeHardLink:
		return tar.TypeLink
	case file.TypeSymLink:
		return tar.TypeSymlink
	case file.TypeCharacterDevice:
		return tar.TypeChar
	case file.TypeBlockDevice:
		return tar.TypeBlock
	case file.TypeDirectory:
		return tar.TypeDir
	case file.TypeFIFO:
		return tar.TypeFifo
	default:
		return tar.TypeReg
	}
}
