package export

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/anchore/stereoscope/pkg/filetree/filenode"

	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/internal/layertree"
	"github.com/wagoodman/dive/internal/log"
)

type Export struct {
	Layer []Layer `json:"layer"`
	Image Image   `json:"image"`
}

type Layer struct {
	Index     int        `json:"index"`
	ID        string     `json:"id"`
	DigestID  string     `json:"digestId"`
	SizeBytes uint64     `json:"sizeBytes"`
	Command   string     `json:"command"`
	FileList  []FileInfo `json:"fileList"`
}

// FileInfo is the tar header metadata of one layer entry.
type FileInfo struct {
	Path     string      `json:"path"`
	TypeFlag byte        `json:"typeFlag"`
	Linkname string      `json:"linkName"`
	Size     int64       `json:"size"`
	Mode     os.FileMode `json:"fileMode"`
	Uid      int         `json:"uid"`
	Gid      int         `json:"gid"`
	IsDir    bool        `json:"isDir"`
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

// NewExport exports an analysis to the JSON shape. Each layer lists the entries of its own tar (whiteouts included,
// opaque whiteout markers and implicit parent directories left out), children before parents.
func NewExport(analysis *dive.Analysis) *Export {
	data := Export{
		Layer: make([]Layer, len(analysis.Layers)),
		Image: Image{
			InefficientFiles: make([]FileReference, len(analysis.Inefficiencies)),
			SizeBytes:        analysis.SizeBytes,
			EfficiencyScore:  analysis.Efficiency,
			InefficientBytes: analysis.WastedBytes,
		},
	}

	img := analysis.Image
	for idx, l := range analysis.Layers {
		fileList := make([]FileInfo, 0)
		err := layertree.Walk(img.Layers[idx].Tree, layertree.OpaqueMarker, func(n *filenode.FileNode, _ bool) error {
			hdr, ok := layertree.Header(img, n.Reference)
			if !ok {
				return nil
			}
			fi := hdr.FileInfo()
			fileList = append(fileList, FileInfo{
				Path:     strings.TrimPrefix(string(n.Reference.RealPath), "/"),
				TypeFlag: hdr.Typeflag,
				Linkname: hdr.Linkname,
				Size:     hdr.Size,
				Mode:     fi.Mode(),
				Uid:      hdr.Uid,
				Gid:      hdr.Gid,
				IsDir:    fi.IsDir(),
			})
			return nil
		})
		if err != nil {
			log.WithFields("layer", l.Digest, "error", err).Debug("unable to walk layer tree")
		}
		data.Layer[idx] = Layer{
			Index:     l.Index,
			ID:        l.Digest,
			DigestID:  l.Digest,
			SizeBytes: l.Size,
			Command:   l.Command,
			FileList:  fileList,
		}
	}

	for idx, ineff := range analysis.Inefficiencies {
		data.Image.InefficientFiles[idx] = FileReference{
			References: ineff.Count,
			SizeBytes:  ineff.SizeBytes,
			Path:       ineff.Path,
		}
	}

	return &data
}

func (exp *Export) Marshal() ([]byte, error) {
	return json.MarshalIndent(&exp, "", "  ")
}
