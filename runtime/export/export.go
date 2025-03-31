package export

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
	diveImage "github.com/wagoodman/dive/dive/image"
)

type export struct {
	Layer []layer `json:"layer"`
	Image image   `json:"image"`
}

// NewExport exports the analysis to a JSON
func NewExport(analysis *diveImage.Analysis) *export {
	data := export{
		Layer: make([]layer, len(analysis.Layers)),
		Image: image{
			InefficientFiles: make([]FileReference, len(analysis.Inefficiencies)),
			SizeBytes:        analysis.SizeBytes,
			EfficiencyScore:  analysis.Efficiency,
			InefficientBytes: analysis.WastedBytes,
		},
	}

	// export layers in order
	for idx, curLayer := range analysis.Layers {
		layerFileList := make([]filetree.FileInfo, 0)
		visitor := func(node *filetree.FileNode) error {
			layerFileList = append(layerFileList, node.Data.FileInfo)
			return nil
		}
		err := curLayer.Tree.VisitDepthChildFirst(visitor, nil)
		if err != nil {
			logrus.Errorf("Unable to propagate layer tree: %+v", err)
		}
		data.Layer[idx] = layer{
			Index:     curLayer.Index,
			ID:        curLayer.Id,
			DigestID:  curLayer.Digest,
			SizeBytes: curLayer.Size,
			Command:   curLayer.Command,
			FileList:  layerFileList,
		}
	}

	// add file references
	for idx := 0; idx < len(analysis.Inefficiencies); idx++ {
		fileData := analysis.Inefficiencies[len(analysis.Inefficiencies)-1-idx]

		data.Image.InefficientFiles[idx] = FileReference{
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
