package filetree

import (
	"fmt"
	"io"
	"strings"

	"github.com/anchore/stereoscope/pkg/file"
	stereofiletree "github.com/anchore/stereoscope/pkg/filetree"
	"github.com/anchore/stereoscope/pkg/filetree/filenode"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/cespare/xxhash/v2"

	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/internal/log"
)

type TreeIndexKey struct {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int
}

func NewTreeIndexKey(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) TreeIndexKey {
	return TreeIndexKey{
		bottomTreeStart: bottomTreeStart,
		bottomTreeStop:  bottomTreeStop,
		topTreeStart:    topTreeStart,
		topTreeStop:     topTreeStop,
	}
}

func (index TreeIndexKey) String() string {
	if index.bottomTreeStart == index.bottomTreeStop && index.topTreeStart == index.topTreeStop {
		return fmt.Sprintf("Index(%d:%d)", index.bottomTreeStart, index.topTreeStart)
	} else if index.bottomTreeStart == index.bottomTreeStop {
		return fmt.Sprintf("Index(%d:%d-%d)", index.bottomTreeStart, index.topTreeStart, index.topTreeStop)
	} else if index.topTreeStart == index.topTreeStop {
		return fmt.Sprintf("Index(%d-%d:%d)", index.bottomTreeStart, index.bottomTreeStop, index.topTreeStart)
	}
	return fmt.Sprintf("Index(%d-%d:%d-%d)", index.bottomTreeStart, index.bottomTreeStop, index.topTreeStart, index.topTreeStop)
}

// Comparer builds marked trees for a pair of layer ranges on demand and caches them. The bottom of a range is the
// squashed tree stereoscope already computed for its top layer, the top layers are each compared in turn.
type Comparer struct {
	img        *image.Image
	layerTrees map[int]*FileTree
	trees      map[TreeIndexKey]*FileTree
	hashes     map[file.ID]uint64
}

func NewComparer(img *image.Image) *Comparer {
	return &Comparer{
		img:        img,
		layerTrees: make(map[int]*FileTree),
		trees:      make(map[TreeIndexKey]*FileTree),
		hashes:     make(map[file.ID]uint64),
	}
}

// GetTree returns the bottom range of layers marked with the changes of the top range. The bottom range always
// starts at layer 0 (bottomTreeStart is ignored, as it was before).
func (cmp *Comparer) GetTree(key TreeIndexKey) (*FileTree, error) {
	if value, exists := cmp.trees[key]; exists {
		return value, nil
	}
	if key.bottomTreeStop < 0 || key.topTreeStop >= len(cmp.img.Layers) || key.bottomTreeStop >= len(cmp.img.Layers) {
		return nil, fmt.Errorf("invalid layer range: %s", key)
	}

	tree := newTree(cmp.img, cmp.img.Layers[key.bottomTreeStop].SquashedTree)
	for idx := key.topTreeStart; idx <= key.topTreeStop; idx++ {
		pathErrors, err := tree.CompareAndMark(cmp.layerTree(idx), cmp.contentHash)
		if err != nil {
			return nil, fmt.Errorf("failed to build tree: %w", err)
		}
		for _, pe := range pathErrors {
			log.WithFields("index", key.String(), "error", pe.String()).Debug("path error while comparing layers")
		}
	}
	cmp.trees[key] = tree
	return tree, nil
}

// layerTree is the UI tree of one layer's own (unsquashed) tree, whiteout entries included
func (cmp *Comparer) layerTree(idx int) *FileTree {
	if t, ok := cmp.layerTrees[idx]; ok {
		return t
	}
	t := newTree(cmp.img, cmp.img.Layers[idx].Tree)
	cmp.layerTrees[idx] = t
	return t
}

// contentHash reads a regular file through the image catalog, once per file reference
func (cmp *Comparer) contentHash(f FileInfo) uint64 {
	if f.ref == nil {
		return f.hash
	}
	if !f.hasContent() {
		return 0
	}
	if h, ok := cmp.hashes[f.ref.ID()]; ok {
		return h
	}
	var h uint64
	rc, err := cmp.img.OpenReference(*f.ref)
	if err == nil {
		d := xxhash.New()
		_, err = io.Copy(d, rc)
		_ = rc.Close()
		h = d.Sum64()
	}
	if err != nil {
		log.WithFields("path", f.ref.RealPath, "error", err).Debug("unable to hash file contents")
	}
	cmp.hashes[f.ref.ID()] = h
	return h
}

// newTree converts a stereoscope tree into a UI tree. Opaque whiteout markers are left out, the old tree never
// added them either.
func newTree(img *image.Image, src stereofiletree.Reader) *FileTree {
	tree := NewFileTree()
	err := dive.WalkTree(src, func(n *filenode.FileNode, _ bool) error {
		if strings.HasPrefix(n.RealPath.Basename(), doubleWhiteoutPrefix) {
			return nil
		}
		_, _, err := tree.AddPath(string(n.RealPath), newFileInfo(img, n.Reference))
		return err
	})
	if err != nil {
		log.WithFields("error", err).Debug("unable to build UI tree")
	}
	return tree
}
