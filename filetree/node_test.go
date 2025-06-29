func NewFileInfo(path string, hash string, mode os.FileMode, size int64) FileInfo {
	return FileInfo{
		Path:     path,
		MD5sum:   hash,
		Mode:     mode,
		Size:     size,
		DiffType: Unmodified,
		XAttrs:   make(map[string][]byte),
	}
}
