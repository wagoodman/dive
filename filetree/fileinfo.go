// FileInfo contains file metadata.
type FileInfo struct {
	Path          string    `json:"path"`
	TypeFlag      byte      `json:"typeFlag"`
	LinkName      string    `json:"linkName"`
	Size          int64     `json:"size"`
	Mode          os.FileMode `json:"mode"`
	UID           int       `json:"uid"`
	GID           int       `json:"gid"`
	ModTime       time.Time `json:"modTime"`
	MD5sum        string    `json:"md5sum"`
	Contents      []byte    `json:"contents"`
	DiffType      DiffType  `json:"diffType"`
	XAttrs        map[string][]byte `json:"xattrs"` // Extended attributes
}
