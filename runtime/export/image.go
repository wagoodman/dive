package export

type image struct {
	SizeBytes        uint64          `json:"sizeBytes"`
	InefficientBytes uint64          `json:"inefficientBytes"`
	EfficiencyScore  float64         `json:"efficiencyScore"`
	InefficientFiles []fileReference `json:"fileReference"`
}
