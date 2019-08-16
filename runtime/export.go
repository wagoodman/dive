package runtime

import (
	"encoding/json"
	"io/ioutil"

	"github.com/wagoodman/dive/image"
)

type export struct {
	Layer []exportLayer `json:"layer"`
	Image exportImage   `json:"image"`
}

type exportLayer struct {
	Index     int    `json:"index"`
	DigestID  string `json:"digestId"`
	SizeBytes uint64 `json:"sizeBytes"`
	Command   string `json:"command"`
}

type exportImage struct {
	SizeBytes        uint64          `json:"sizeBytes"`
	InefficientBytes uint64          `json:"inefficientBytes"`
	EfficiencyScore  float64         `json:"efficiencyScore"`
	InefficientFiles []ReferenceFile `json:"ReferenceFile"`
}

func newExport(analysis *image.AnalysisResult) *export {
	data := export{}
	data.Layer = make([]exportLayer, len(analysis.Layers))
	data.Image.InefficientFiles = make([]ReferenceFile, len(analysis.Inefficiencies))

	// export layers in order
	for revIdx := len(analysis.Layers) - 1; revIdx >= 0; revIdx-- {
		layer := analysis.Layers[revIdx]
		idx := (len(analysis.Layers) - 1) - revIdx

		data.Layer[idx] = exportLayer{
			Index:     layer.Index(),
			DigestID:  layer.Id(),
			SizeBytes: layer.Size(),
			Command:   layer.Command(),
		}
	}

	data.Image.SizeBytes = analysis.SizeBytes
	data.Image.EfficiencyScore = analysis.Efficiency
	data.Image.InefficientBytes = analysis.WastedBytes

	for idx := 0; idx < len(analysis.Inefficiencies); idx++ {
		fileData := analysis.Inefficiencies[len(analysis.Inefficiencies)-1-idx]

		data.Image.InefficientFiles[idx] = ReferenceFile{
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

func (exp *export) toFile(exportFilePath string) error {
	payload, err := exp.marshal()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(exportFilePath, payload, 0644)
}
