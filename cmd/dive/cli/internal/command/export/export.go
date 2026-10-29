package export

import (
	"archive/tar"
	"encoding/json"
	"os"
	"strings"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/filetree/filenode"

	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/dive/filetree"
	diveImage "github.com/wagoodman/dive/dive/image"
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
		layerFileList := make([]FileInfo, 0)
		visitor := func(node *filetree.FileNode) error {
			fi := node.Data.FileInfo
			layerFileList = append(layerFileList, FileInfo{
				Path:     fi.Path,
				TypeFlag: fi.TypeFlag,
				Linkname: fi.Linkname,
				Size:     fi.Size,
				Mode:     fi.Mode,
				Uid:      fi.Uid,
				Gid:      fi.Gid,
				IsDir:    fi.IsDir,
			})
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

// NewExportFromAnalysis exports a stereoscope based analysis, in the same shape as NewExport. Each layer lists the
// entries of its own tar (whiteouts included, implicit parent directories left out) children before parents.
func NewExportFromAnalysis(analysis *dive.Analysis) *Export {
	data := Export{
		Layer: make([]Layer, len(analysis.Layers)),
		Image: Image{
			InefficientFiles: make([]FileReference, len(analysis.Inefficiencies)),
			SizeBytes:        analysis.SizeBytes,
			EfficiencyScore:  analysis.Efficiency,
			InefficientBytes: analysis.WastedBytes,
		},
	}

	catalog := analysis.Image.FileCatalog
	for idx, l := range analysis.Layers {
		fileList := make([]FileInfo, 0)
		err := dive.WalkTree(analysis.Image.Layers[idx].Tree, func(n *filenode.FileNode, _ bool) error {
			if n.Reference == nil {
				return nil
			}
			entry, err := catalog.Get(*n.Reference)
			if err != nil {
				return nil
			}
			fi := FileInfo{
				Path:     strings.TrimPrefix(entry.Path, "/"),
				TypeFlag: tarTypeFlag(entry.Type),
				Linkname: entry.LinkDestination,
				Size:     entry.Size(),
				Mode:     entry.Mode(),
				Uid:      entry.UserID,
				Gid:      entry.GroupID,
				IsDir:    entry.IsDir(),
			}
			// prefer the header as written: stereoscope describes an in-layer hardlink as the file it names
			if hdr := dive.TarHeader(entry.Metadata); hdr != nil {
				fi.TypeFlag = hdr.Typeflag
				fi.Linkname = hdr.Linkname
				fi.Size = hdr.Size
				fi.Mode = hdr.FileInfo().Mode()
			}
			fileList = append(fileList, fi)
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

// tarTypeFlag maps a stereoscope file type back to the tar typeflag it was read from.
func tarTypeFlag(t file.Type) byte {
	switch t {
	case file.TypeHardLink:
		return tar.TypeLink
	case file.TypeSymLink:
		return tar.TypeSymlink
	case file.TypeCharacterDevice:
		return tar.TypeChar
	case file.TypeBlockDevice:
		return tar.TypeBlock
	case file.TypeDirectory:
		return tar.TypeDir
	case file.TypeFIFO:
		return tar.TypeFifo
	default:
		return tar.TypeReg
	}
}

func (exp *Export) Marshal() ([]byte, error) {
	return json.MarshalIndent(&exp, "", "  ")
}
