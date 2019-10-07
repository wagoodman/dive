package export

import (
	"encoding/json"
	diveImage "github.com/wagoodman/dive/dive/image"
	"io/ioutil"
)

type export struct {
	Layer []layer `json:"layer"`
	Image image   `json:"image"`
}

func NewExport(analysis *diveImage.AnalysisResult) *export {
	data := export{}
	data.Layer = make([]layer, len(analysis.Layers))
	data.Image.InefficientFiles = make([]fileReference, len(analysis.Inefficiencies))

	// export layers in order
	for revIdx := len(analysis.Layers) - 1; revIdx >= 0; revIdx-- {
		curLayer := analysis.Layers[revIdx]
		idx := (len(analysis.Layers) - 1) - revIdx

		data.Layer[idx] = layer{
			Index:     curLayer.Index,
			DigestID:  curLayer.Id,
			SizeBytes: curLayer.Size,
			Command:   curLayer.Command,
		}
	}

	data.Image.SizeBytes = analysis.SizeBytes
	data.Image.EfficiencyScore = analysis.Efficiency
	data.Image.InefficientBytes = analysis.WastedBytes

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

func (exp *export) marshal() ([]byte, error) {
	return json.MarshalIndent(&exp, "", "  ")
}

func (exp *export) ToFile(exportFilePath string) error {
	payload, err := exp.marshal()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(exportFilePath, payload, 0644)
}
