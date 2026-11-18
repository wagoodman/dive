package filetree

import (
	"archive/tar"
	"maps"
	"os"
	"strings"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/image"

	"github.com/wagoodman/dive/internal/layertree"
)

// FileInfo contains tar metadata for a specific FileNode
type FileInfo struct {
	Path     string
	TypeFlag byte
	Linkname string
	Size     int64
	Mode     os.FileMode
	Uid      int
	Gid      int
	IsDir    bool
	// xattrs are the extended attributes (e.g. file capabilities) from the tar PAX records, nil when there are none
	xattrs map[string]string
	// implicit marks a dir that has no entry of its own, it only exists as the parent of other entries
	implicit bool
	// ref points at the file in the image catalog, it is how content is read (and hashed) lazily
	ref *file.Reference
}

// newFileInfo describes a stereoscope file from its original tar header, so a hardlink is still a hardlink (size 0,
// its own link name) and not the file it names. A node without a catalog entry is an implicit parent dir, it gets
// the 0755 root owned dir a runtime would create for it.
func newFileInfo(img *image.Image, ref *file.Reference) FileInfo {
	hdr, ok := layertree.Header(img, ref)
	if !ok {
		return FileInfo{TypeFlag: tar.TypeDir, Mode: os.ModeDir | 0o755, IsDir: true, implicit: true}
	}
	fi := hdr.FileInfo()
	return FileInfo{
		Path:     strings.TrimPrefix(string(ref.RealPath), "/"),
		TypeFlag: hdr.Typeflag,
		Linkname: hdr.Linkname,
		Size:     fi.Size(),
		Mode:     fi.Mode(),
		Uid:      hdr.Uid,
		Gid:      hdr.Gid,
		IsDir:    fi.IsDir(),
		xattrs:   xattrs(hdr),
		ref:      ref,
	}
}

// xattrs picks the extended attributes out of a tar header's PAX records
func xattrs(hdr *tar.Header) map[string]string {
	var x map[string]string
	for k, v := range hdr.PAXRecords {
		if name, ok := strings.CutPrefix(k, "SCHILY.xattr."); ok {
			if x == nil {
				x = make(map[string]string)
			}
			x[name] = v
		}
	}
	return x
}

// Copy duplicates a FileInfo
func (data *FileInfo) Copy() *FileInfo {
	if data == nil {
		return nil
	}
	c := *data
	return &c
}

// hasContent reports whether the file has bytes worth hashing. Everything else (dirs, links, devices) has an empty
// data section in the tar, so two of the same type always have equal content (a hardlink is not read through to
// the file it names, stereoscope would do that). TypeRegA is not checked: Go's tar reader rewrites it to TypeReg.
func (data *FileInfo) hasContent() bool {
	return data.TypeFlag == tar.TypeReg
}

// ContentHasher returns the content hash of a file, reporting false when the content could not be read
type ContentHasher func(FileInfo) (uint64, bool)

// Compare determines the DiffType between two FileInfos based on the type and contents of each given FileInfo.
// Content is only hashed when everything else matches, and an unreadable side counts as Modified. A nil hasher
// compares metadata only.
func (data *FileInfo) Compare(other FileInfo, hash ContentHasher) DiffType {
	// the same catalog entry on both sides (a layer against itself) can't differ, don't read it
	if data.ref != nil && other.ref != nil && data.ref.ID() == other.ref.ID() {
		return Unmodified
	}
	// a layer that only implies a dir (no entry of its own) doesn't change the dir below it
	if other.implicit && data.IsDir {
		return Unmodified
	}
	if data.TypeFlag != other.TypeFlag ||
		data.Mode != other.Mode ||
		data.Uid != other.Uid ||
		data.Gid != other.Gid ||
		data.Size != other.Size ||
		data.Linkname != other.Linkname ||
		!maps.Equal(data.xattrs, other.xattrs) {
		return Modified
	}
	if hash == nil || !data.hasContent() {
		return Unmodified
	}
	a, okA := hash(*data)
	b, okB := hash(other)
	if okA && okB && a == b {
		return Unmodified
	}
	return Modified
}
