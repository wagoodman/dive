// NewFileInfoFromTarHeader creates a FileInfo object from a tar header and file contents
func NewFileInfoFromTarHeader(reader io.Reader, header *tar.Header, name string) FileInfo {
	if len(name) == 0 {
		name = header.Name
	}
	var contents []byte
	if reader != nil && header.Size > 0 {
		contents, _ = io.ReadAll(reader)
	}

	contentHash := ""
	if len(contents) > 0 {
		hashBytes := md5.Sum(contents)
		contentHash = hex.EncodeToString(hashBytes[:])
	}

	return FileInfo{
		Path:          name,
		TypeFlag:      header.Typeflag,
		LinkName:      header.Linkname,
		Size:          header.Size,
		Mode:          header.FileInfo().Mode(),
		UID:           header.Uid,
		GID:           header.Gid,
		ModTime:       header.ModTime,
		MD5sum:        contentHash,
		Contents:      contents,
		XAttrs:        make(map[string][]byte),
	}
}
