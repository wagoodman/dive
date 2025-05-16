package export

import (
	"encoding/json"
	"github.com/wagoodman/dive/dive/v1/filetree"
	diveImage "github.com/wagoodman/dive/dive/v1/image"
	"github.com/wagoodman/dive/internal/log"
)

type Export struct {
	Layer []Layer `json:"layer"`
	Image Image   `json:"image"`
}

type Layer struct {
	Index     int                 `json:"index"`
	ID        string              `json:"id"`
	DigestID  string              `json:"digestId"`
	SizeBytes uint64              `json:"sizeBytes"`
	Command   string              `json:"command"`
	FileList  []filetree.FileInfo `json:"fileList"`
}

type Image struct {
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

// NewExport exports the analysis to a JSON
func NewExport(analysis *diveImage.Analysis) *Export {
	data := Export{
		Layer: make([]Layer, len(analysis.Layers)),
		Image: Image{
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
			log.WithFields("layer", curLayer.Id, "error", err).Debug("unable to propagate layer tree")
		}
		data.Layer[idx] = Layer{
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

func (exp *Export) Marshal() ([]byte, error) {
	return json.MarshalIndent(&exp, "", "  ")
}
