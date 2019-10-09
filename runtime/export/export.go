package export

import (
	"encoding/json"
	diveImage "github.com/wagoodman/dive/dive/image"
)

type export struct {
	Layer []layer `json:"layer"`
	Image image   `json:"image"`
}

func NewExport(analysis *diveImage.AnalysisResult) *export {
	data := export{
		Layer: make([]layer, len(analysis.Layers)),
		Image: image{
			InefficientFiles: make([]fileReference, len(analysis.Inefficiencies)),
			SizeBytes:        analysis.SizeBytes,
			EfficiencyScore:  analysis.Efficiency,
			InefficientBytes: analysis.WastedBytes,
		},
	}

	// export layers in order
	for idx, curLayer := range analysis.Layers {
		data.Layer[idx] = layer{
			Index:     curLayer.Index,
			ID:        curLayer.Id,
			DigestID:  curLayer.Digest,
			SizeBytes: curLayer.Size,
			Command:   curLayer.Command,
		}
	}

	// add file references
	for idx := 0; idx < len(analysis.Inefficiencies); idx++ {
		fileData := analysis.Inefficiencies[len(analysis.Inefficiencies)-1-idx]

		data.Image.InefficientFiles[idx] = fileReference{
			References: len(fileData.Nodes),
			SizeBytes:  uint64(fileData.CumulativeSize),
			Path:       fileData.Path,
		}
	}

	return &data
}

func (exp *export) Marshal() ([]byte, error) {
	return json.MarshalIndent(&exp, "", "  ")
}
