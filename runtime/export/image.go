package export

type image struct {
	SizeBytes        uint64          `json:"sizeBytes"`
	InefficientBytes uint64          `json:"inefficientBytes"`
	EfficiencyScore  float64         `json:"efficiencyScore"`
	InefficientFiles []FileReference `json:"fileReference"`
}

type FileReference struct {
	References int    `json:"count"`
	SizeBytes  uint64 `json:"sizeBytes"`
	Path       string `json:"file"`
}
