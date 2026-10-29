package filetree

import (
	"archive/tar"
	"os"
	"strings"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/image"

	"github.com/wagoodman/dive/dive"
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
	// hash is a fixed content hash, only used when there is no ref (tests)
	hash uint64
}

// newFileInfo describes a stereoscope file from its original tar header, so a hardlink is still a hardlink (size 0,
// its own link name) and not the file it names. A node without a catalog entry (an implicit parent dir) gets an
// empty FileInfo, the same as the old tree gave a dir that had no tar entry.
func newFileInfo(img *image.Image, ref *file.Reference) FileInfo {
	if ref == nil {
		return FileInfo{}
	}
	entry, err := img.FileCatalog.Get(*ref)
	if err != nil {
		return FileInfo{}
	}
	hdr := dive.TarHeader(entry.Metadata)
	if hdr == nil {
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
// the file it names, stereoscope would do that).
func (data *FileInfo) hasContent() bool {
	//nolint:staticcheck // old tars can still carry TypeRegA
	return data.TypeFlag == tar.TypeReg || data.TypeFlag == tar.TypeRegA
}

// ContentHasher returns the content hash of a file
type ContentHasher func(FileInfo) uint64

func fixedHash(f FileInfo) uint64 { return f.hash }

// Compare determines the DiffType between two FileInfos based on the type and contents of each given FileInfo.
// Content is only hashed when everything else matches.
func (data *FileInfo) Compare(other FileInfo, hash ContentHasher) DiffType {
	if hash == nil {
		hash = fixedHash
	}
	if data.TypeFlag == other.TypeFlag &&
		data.Mode == other.Mode &&
		data.Uid == other.Uid &&
		data.Gid == other.Gid {
		if hash(*data) == hash(other) {
			return Unmodified
		}
	}
	return Modified
}
