package filetree

import (
	"archive/tar"
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
	// ref points at the file in the image catalog, it is how content is read (and hashed) lazily
	ref *file.Reference
}

// newFileInfo describes a stereoscope file from its original tar header, so a hardlink is still a hardlink (size 0,
// its own link name) and not the file it names. A node without a catalog entry (an implicit parent dir) gets an
// empty FileInfo, the same as the old tree gave a dir that had no tar entry.
func newFileInfo(img *image.Image, ref *file.Reference) FileInfo {
	hdr, ok := layertree.Header(img, ref)
	if !ok {
		return FileInfo{}
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
		ref:      ref,
	}
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
// the file it names, stereoscope would do that). TypeRegA is not checked: Go's tar reader rewrites it to TypeReg,
// so a zero TypeFlag only ever means the empty FileInfo of an implicit parent dir.
func (data *FileInfo) hasContent() bool {
	return data.TypeFlag == tar.TypeReg
}

// ContentHasher returns the content hash of a file, reporting false when the content could not be read
type ContentHasher func(FileInfo) (uint64, bool)

// Compare determines the DiffType between two FileInfos based on the type and contents of each given FileInfo.
// Content is only hashed when everything else matches, and an unreadable side counts as Modified. A nil hasher
// compares metadata only.
func (data *FileInfo) Compare(other FileInfo, hash ContentHasher) DiffType {
	if data.TypeFlag != other.TypeFlag ||
		data.Mode != other.Mode ||
		data.Uid != other.Uid ||
		data.Gid != other.Gid ||
		data.Size != other.Size ||
		data.Linkname != other.Linkname {
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
