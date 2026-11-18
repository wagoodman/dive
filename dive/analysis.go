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

// Inefficiency is a path that wastes bytes because a later layer overwrote or removed it.
type Inefficiency struct {
	Path string
	// Count is how many layers the path shows up in (added or removed)
	Count int
	// SizeBytes is the bytes wasted on the path: every copy layers added minus the copy left in the final image
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
	// count is how many layers add or remove the path
	count int
	// cumulative is the size of every copy layers added
	cumulative int64
	// final is the size of the copy the final image holds, 0 once removed
	final int64
}

// imageEfficiency scores the image and lists the paths that waste bytes.
//
// waste is defined per path as cumulative - final: every byte layers added for the path minus the bytes the final
// image still holds there. An overwritten copy, a removed file and a file removed and added back again each waste
// exactly the bytes that are stuck in lower layers yet unreachable in the image. The score is the share of all
// added bytes that are not wasted.
//
// It walks the leaf paths of every layer's own tree. A path a layer removes is an occurrence that sets its final
// size to 0. Removal is anything that makes a path seen in a lower layer disappear from this layer's squashed tree:
// a whiteout of the file or of a parent dir, an opaque dir, or a file or link put in place of a parent dir.
func imageEfficiency(ctx context.Context, img *image.Image) (float64, []Inefficiency, error) {
	byPath := make(map[file.Path]*pathEfficiency)
	var order []file.Path

	record := func(p file.Path, size int64) {
		data, ok := byPath[p]
		if !ok {
			data = &pathEfficiency{}
			byPath[p] = data
			order = append(order, p)
		}
		data.count++
		data.cumulative += size
		data.final = size
	}

	for i, l := range img.Layers {
		if err := ctx.Err(); err != nil {
			return 0, nil, err
		}
		// paths of this layer that may hide lower paths: whiteout targets, opaque dirs and non-dirs
		var roots []file.Path
		err := layertree.Walk(l.Tree, nil, func(n *filenode.FileNode, leaf bool) error {
			p := n.RealPath
			if p.IsWhiteout() {
				// for an opaque marker this is the dir it sits in
				target, err := p.UnWhiteoutPath()
				if err != nil {
					return err
				}
				roots = append(roots, target)
				return nil
			}
			if !leaf {
				return nil
			}
			if n.FileType != file.TypeDirectory {
				roots = append(roots, p)
			}
			record(p, fileSize(img, n.Reference))
			return nil
		})
		if err != nil {
			return 0, nil, fmt.Errorf("unable to walk layer %d: %w", i, err)
		}
		if i == 0 {
			continue
		}
		for _, p := range removedPaths(img.Layers[i-1].SquashedTree, l.SquashedTree, roots) {
			if data, ok := byPath[p]; ok {
				data.count++
				data.final = 0
			}
		}
	}

	var wasted, discovered int64
	var inefficiencies []Inefficiency
	for _, p := range order {
		data := byPath[p]
		waste := data.cumulative - data.final
		wasted += waste
		discovered += data.cumulative
		if waste > 0 {
			inefficiencies = append(inefficiencies, Inefficiency{Path: string(p), Count: data.count, SizeBytes: uint64(waste)})
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
		score = float64(discovered-wasted) / float64(discovered)
	}
	return score, inefficiencies, nil
}

// removedPaths lists the paths at or under roots in the lower squashed tree that are gone from the upper one.
func removedPaths(lower, upper filetree.Reader, roots []file.Path) []file.Path {
	lt, ut := lower.TreeReader(), upper.TreeReader()
	seen := make(map[file.Path]bool)
	var removed []file.Path
	var visit func(node.Node)
	visit = func(n node.Node) {
		p := n.(*filenode.FileNode).RealPath
		if seen[p] {
			return
		}
		seen[p] = true
		if ut.Node(n.ID()) == nil {
			removed = append(removed, p)
		}
		for _, child := range lt.Children(n) {
			visit(child)
		}
	}
	for _, root := range roots {
		if n := lt.Node(filenode.IDByPath(root)); n != nil {
			visit(n)
		}
	}
	return removed
}

// fileSize is the size the tar header recorded for ref, so a hardlink is 0 bytes (stereoscope describes an
// in-layer hardlink as the file it names, with that file's size).
func fileSize(img *image.Image, ref *file.Reference) int64 {
	if hdr, ok := layertree.Header(img, ref); ok {
		return hdr.Size
	}
	return 0
}
