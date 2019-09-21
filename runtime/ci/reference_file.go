package ci

type ReferenceFile struct {
	References int    `json:"count"`
	SizeBytes  uint64 `json:"sizeBytes"`
	Path       string `json:"file"`
}
