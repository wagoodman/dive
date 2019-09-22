package filetree

import (
	"archive/tar"
	"github.com/cespare/xxhash"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

// FileInfo contains tar metadata for a specific FileNode
type FileInfo struct {
	Path     string
	TypeFlag byte
	Linkname string
	hash     uint64
	Size     int64
	Mode     os.FileMode
	Uid      int
	Gid      int
	IsDir    bool
}

// NewFileInfo extracts the metadata from a tar header and file contents and generates a new FileInfo object.
func NewFileInfo(reader *tar.Reader, header *tar.Header, path string) FileInfo {
	if header.Typeflag == tar.TypeDir {
		return FileInfo{
			Path:     path,
			TypeFlag: header.Typeflag,
			Linkname: header.Linkname,
			hash:     0,
			Size:     header.FileInfo().Size(),
			Mode:     header.FileInfo().Mode(),
			Uid:      header.Uid,
			Gid:      header.Gid,
			IsDir:    header.FileInfo().IsDir(),
		}
	}

	hash := getHashFromReader(reader)

	return FileInfo{
		Path:     path,
		TypeFlag: header.Typeflag,
		Linkname: header.Linkname,
		hash:     hash,
		Size:     header.FileInfo().Size(),
		Mode:     header.FileInfo().Mode(),
		Uid:      header.Uid,
		Gid:      header.Gid,
		IsDir:    header.FileInfo().IsDir(),
	}
}

// Copy duplicates a FileInfo
func (data *FileInfo) Copy() *FileInfo {
	if data == nil {
		return nil
	}
	return &FileInfo{
		Path:     data.Path,
		TypeFlag: data.TypeFlag,
		Linkname: data.Linkname,
		hash:     data.hash,
		Size:     data.Size,
		Mode:     data.Mode,
		Uid:      data.Uid,
		Gid:      data.Gid,
		IsDir:    data.IsDir,
	}
}

// Compare determines the DiffType between two FileInfos based on the type and contents of each given FileInfo
func (data *FileInfo) Compare(other FileInfo) DiffType {
	if data.TypeFlag == other.TypeFlag {
		if data.hash == other.hash &&
			data.Mode == other.Mode &&
			data.Uid == other.Uid &&
			data.Gid == other.Gid {
			return Unmodified
		}
	}
	return Modified
}

func getHashFromReader(reader io.Reader) uint64 {
	h := xxhash.New()

	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			logrus.Panic(err)
		}
		if n == 0 {
			break
		}

		_, err = h.Write(buf[:n])
		if err != nil {
			logrus.Panic(err)
		}
	}

	return h.Sum64()
}
