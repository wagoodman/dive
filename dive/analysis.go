package dive

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/filetree"
	"github.com/anchore/stereoscope/pkg/filetree/filenode"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/anchore/stereoscope/pkg/tree/node"

	"github.com/wagoodman/dive/internal/layertree"
)

// Analysis is the result of analyzing an image: the stereoscope image itself plus layer and efficiency stats.
//
// The image is borrowed, not owned: it is the caller's image, and anything that reads file contents through it
// (Image.FileCatalog, Image.OpenReference) only works until the caller runs Image.Cleanup().
type Analysis struct {
	// Ref is a label for the image, as passed to Analyze
	Ref   string
	Image *image.Image
	// Layers is in build order, Layers[i] describes Image.Layers[i]
	Layers []Layer
	// Efficiency is 1.0 for an image without duplicated or removed files, lower otherwise
	Efficiency float64
	// SizeBytes is the sum of all layer sizes
	SizeBytes uint64
	// UserSizeBytes is the sum of all layer sizes except the first (base) layer
	UserSizeBytes uint64
	// WastedBytes is the sum of Inefficiencies sizes
	WastedBytes uint64
	// WastedUserPercent is WastedBytes / UserSizeBytes, 0 when there are no user bytes
	WastedUserPercent float64
	// Inefficiencies is sorted by size, largest first, ties by path
	Inefficiencies []Inefficiency
}

// Layer is the per layer info of an Analysis.
type Layer struct {
	Index int
	// Digest is the layer diff ID
	Digest string
	// Command is the history entry that created the layer, "(missing)" when the image has no history for it
	Command string
	Size    uint64
}

// Inefficiency is a path that shows up in more than one layer (duplicated, overwritten or removed).
type Inefficiency struct {
	Path string
	// Count is how many layers the path shows up in (as a file or a whiteout)
	Count int
	// SizeBytes is the cumulative size of every occurrence
	SizeBytes uint64
}

// Analyze derives layer info and efficiency stats from an already read stereoscope image. ref only labels the result.
func Analyze(ctx context.Context, ref string, img *image.Image) (*Analysis, error) {
	if img == nil {
		return nil, fmt.Errorf("no image to analyze")
	}

	layers := layerInfo(img)
	efficiency, inefficiencies, err := imageEfficiency(ctx, img)
	if err != nil {
		return nil, err
	}

	var sizeBytes, userSizeBytes, wastedBytes uint64
	for i, l := range layers {
		sizeBytes += l.Size
		if i != 0 {
			userSizeBytes += l.Size
		}
	}
	for _, ineff := range inefficiencies {
		wastedBytes += ineff.SizeBytes
	}

	var wastedUserPercent float64
	if userSizeBytes != 0 {
		wastedUserPercent = float64(wastedBytes) / float64(userSizeBytes)
	}

	return &Analysis{
		Ref:               ref,
		Image:             img,
		Layers:            layers,
		Efficiency:        efficiency,
		SizeBytes:         sizeBytes,
		UserSizeBytes:     userSizeBytes,
		WastedBytes:       wastedBytes,
		WastedUserPercent: wastedUserPercent,
		Inefficiencies:    inefficiencies,
	}, nil
}

// layerInfo pairs layers with history entries. History is in build order and has entries for metadata only
// instructions (empty_layer), which own no layer and are skipped. History is optional, so a layer without an
// entry gets "(missing)".
func layerInfo(img *image.Image) []Layer {
	history := img.Metadata.Config.History
	layers := make([]Layer, len(img.Layers))
	histIdx := 0
	for idx, l := range img.Layers {
		createdBy := "(missing)"
		for histIdx < len(history) && history[histIdx].EmptyLayer {
			histIdx++
		}
		if histIdx < len(history) {
			createdBy = history[histIdx].CreatedBy
			histIdx++
		}
		layers[idx] = Layer{
			Index:   int(l.Metadata.Index),
			Digest:  l.Metadata.Digest,
			Command: strings.TrimPrefix(createdBy, "/bin/sh -c "),
			Size:    uint64(l.Metadata.Size),
		}
	}
	return layers
}

type pathEfficiency struct {
	count      int
	cumulative int64
	min        int64
}

// imageEfficiency scores the image and lists paths found in more than one layer. Loosely:
//  1. files duplicated across layers discount the score, weighted by file size
//  2. removed files discount the score, weighted by the original file size
//
// It walks the leaf paths of every layer's own tree. A whiteout counts as an occurrence of the path it removes,
// sized from the squashed tree of the layers below it: a removed directory is the size of everything under it,
// a removed file is 0 (its earlier occurrence already carries the size). Opaque whiteout markers are ignored, so
// the lower files they hide are not counted as removed.
func imageEfficiency(ctx context.Context, img *image.Image) (float64, []Inefficiency, error) {
	byPath := make(map[file.Path]*pathEfficiency)
	var order []file.Path

	for i, l := range img.Layers {
		if err := ctx.Err(); err != nil {
			return 0, nil, err
		}
		err := layertree.Walk(l.Tree, layertree.OpaqueMarker, func(n *filenode.FileNode, leaf bool) error {
			if !leaf {
				return nil
			}
			p := n.RealPath
			var size int64
			if p.IsWhiteout() {
				target, err := p.UnWhiteoutPath()
				if err != nil {
					return err
				}
				var found bool
				if i > 0 {
					size, found = removedSize(img, img.Layers[i-1].SquashedTree, target)
				}
				if !found {
					// nothing below this layer has the path, so the whiteout removes nothing
					return nil
				}
				p = target
			} else {
				size = fileSize(img, n.Reference)
			}

			data, ok := byPath[p]
			if !ok {
				data = &pathEfficiency{min: size}
				byPath[p] = data
				order = append(order, p)
			}
			data.count++
			data.cumulative += size
			data.min = min(data.min, size)
			return nil
		})
		if err != nil {
			return 0, nil, fmt.Errorf("unable to walk layer %d: %w", i, err)
		}
	}

	var minimum, discovered int64
	var inefficiencies []Inefficiency
	for _, p := range order {
		data := byPath[p]
		minimum += data.min
		discovered += data.cumulative
		if data.count > 1 {
			inefficiencies = append(inefficiencies, Inefficiency{Path: string(p), Count: data.count, SizeBytes: uint64(data.cumulative)})
		}
	}

	sort.Slice(inefficiencies, func(i, j int) bool {
		a, b := inefficiencies[i], inefficiencies[j]
		if a.SizeBytes != b.SizeBytes {
			return a.SizeBytes > b.SizeBytes
		}
		return a.Path < b.Path
	})

	score := 1.0
	if discovered != 0 {
		score = float64(minimum) / float64(discovered)
	}
	return score, inefficiencies, nil
}

// removedSize sizes what a whiteout of target removes from the lower squashed tree, reporting false when the lower
// tree has no such path. Only directories have a size here: a removed file's bytes are already counted by the layer
// that added it.
func removedSize(img *image.Image, lower filetree.Reader, target file.Path) (int64, bool) {
	t := lower.TreeReader()
	n := t.Node(filenode.IDByPath(target))
	if n == nil {
		return 0, false
	}
	if n.(*filenode.FileNode).FileType != file.TypeDirectory {
		return 0, true
	}
	var size int64
	var sum func(node.Node)
	sum = func(cur node.Node) {
		size += fileSize(img, cur.(*filenode.FileNode).Reference)
		for _, child := range t.Children(cur) {
			sum(child)
		}
	}
	sum(n)
	return size, true
}

// fileSize is the size the tar header recorded for ref, so a hardlink is 0 bytes (stereoscope describes an
// in-layer hardlink as the file it names, with that file's size).
func fileSize(img *image.Image, ref *file.Reference) int64 {
	if hdr, ok := layertree.Header(img, ref); ok {
		return hdr.Size
	}
	return 0
}
