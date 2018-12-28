package runtime

import (
	"encoding/json"
	"github.com/wagoodman/dive/image"
	"io/ioutil"
)

func newExport(analysis *image.AnalysisResult) *export {
	data := export{}
	data.Layer = make([]exportLayer, len(analysis.Layers))
	data.Image.InefficientFiles = make([]inefficientFiles, len(analysis.Inefficiencies))

	// export layers in order
	for revIdx := len(analysis.Layers) - 1; revIdx >= 0; revIdx-- {
		layer := analysis.Layers[revIdx]
		idx := (len(analysis.Layers) - 1) - revIdx

		data.Layer[idx] = exportLayer{
			Index:     idx,
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

		data.Image.InefficientFiles[idx] = inefficientFiles{
			Count:     len(fileData.Nodes),
			SizeBytes: uint64(fileData.CumulativeSize),
			File:      fileData.Path,
		}
	}

	return &data
}

func exportResultsToFile(analysis *image.AnalysisResult, exportFilePath string) {
	data := newExport(analysis)
	payload, err := json.MarshalIndent(&data, "", "  ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(exportFilePath, payload, 0644)
	if err != nil {
		panic(err)
	}
}