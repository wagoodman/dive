package filetree

import (
	"fmt"
	"io"
	"path"

	"github.com/anchore/stereoscope/pkg/file"
	stereofiletree "github.com/anchore/stereoscope/pkg/filetree"
	"github.com/anchore/stereoscope/pkg/filetree/filenode"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/cespare/xxhash/v2"

	"github.com/wagoodman/dive/internal/layertree"
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
	layerTrees map[int]layerTree
	trees      map[TreeIndexKey]*FileTree
	hashes     map[file.ID]uint64
}

// layerTree is the UI tree of one layer's own (unsquashed) tree, whiteout entries included, plus the directories
// the layer made opaque
type layerTree struct {
	tree       *FileTree
	opaqueDirs []string
}

func NewComparer(img *image.Image) *Comparer {
	return &Comparer{
		img:        img,
		layerTrees: make(map[int]layerTree),
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

	tree, _ := newTree(cmp.img, cmp.img.Layers[key.bottomTreeStop].SquashedTree)
	for idx := key.topTreeStart; idx <= key.topTreeStop; idx++ {
		top := cmp.layerTree(idx)
		pathErrors, err := tree.CompareAndMark(top.tree, cmp.contentHash)
		if err != nil {
			return nil, fmt.Errorf("failed to build tree: %w", err)
		}
		for _, pe := range pathErrors {
			log.WithFields("index", key.String(), "error", pe.String()).Debug("path error while comparing layers")
		}
		for _, dir := range top.opaqueDirs {
			if err := markHidden(tree, top.tree, dir); err != nil {
				return nil, fmt.Errorf("failed to build tree: %w", err)
			}
		}
	}
	cmp.trees[key] = tree
	return tree, nil
}

func (cmp *Comparer) layerTree(idx int) layerTree {
	if t, ok := cmp.layerTrees[idx]; ok {
		return t
	}
	tree, opaqueDirs := newTree(cmp.img, cmp.img.Layers[idx].Tree)
	t := layerTree{tree: tree, opaqueDirs: opaqueDirs}
	cmp.layerTrees[idx] = t
	return t
}

// markHidden marks what an opaque dir hides as Removed: everything the lower tree has under dir that the upper
// layer doesn't have itself.
func markHidden(lower, upper *FileTree, dir string) error {
	lowerDir, err := lower.GetNode(dir)
	if err != nil {
		// nothing below had the dir
		return nil
	}
	upperDir, _ := upper.GetNode(dir)

	var hide func(lowerNode, upperNode *FileNode) (bool, error)
	hide = func(lowerNode, upperNode *FileNode) (bool, error) {
		var hidden bool
		for name, child := range lowerNode.Children {
			var upperChild *FileNode
			if upperNode != nil {
				upperChild = upperNode.Children[name]
			}
			if upperChild == nil {
				if err := child.AssignDiffType(Removed); err != nil {
					return false, err
				}
				hidden = true
				continue
			}
			h, err := hide(child, upperChild)
			if err != nil {
				return false, err
			}
			if h && child.Data.DiffType == Unmodified {
				child.Data.DiffType = Modified
			}
			hidden = hidden || h
		}
		return hidden, nil
	}

	hidden, err := hide(lowerDir, upperDir)
	if err != nil || !hidden {
		return err
	}
	// the dir and its ancestors changed, as a plain whiteout would mark them (the root is never marked)
	for n := lowerDir; n.Parent != nil; n = n.Parent {
		if n.Data.DiffType == Unmodified {
			n.Data.DiffType = Modified
		}
	}
	return nil
}

// contentHash reads a regular file through the image catalog, once per file reference. A read that fails is not
// cached, so it is retried the next time the file is compared.
func (cmp *Comparer) contentHash(f FileInfo) (uint64, bool) {
	if f.ref == nil {
		return 0, false
	}
	if h, ok := cmp.hashes[f.ref.ID()]; ok {
		return h, true
	}
	rc, err := cmp.img.OpenReference(*f.ref)
	if err != nil {
		log.WithFields("path", f.ref.RealPath, "error", err).Debug("unable to read file contents")
		return 0, false
	}
	defer rc.Close()
	d := xxhash.New()
	if _, err := io.Copy(d, rc); err != nil {
		log.WithFields("path", f.ref.RealPath, "error", err).Debug("unable to hash file contents")
		return 0, false
	}
	h := d.Sum64()
	cmp.hashes[f.ref.ID()] = h
	return h, true
}

// newTree converts a stereoscope tree into a UI tree and lists the dirs that opaque whiteout markers made opaque.
// The markers themselves are not added, the old tree never had them either.
func newTree(img *image.Image, src stereofiletree.Reader) (*FileTree, []string) {
	tree := NewFileTree()
	var opaqueDirs []string
	err := layertree.Walk(src, nil, func(n *filenode.FileNode, _ bool) error {
		if layertree.OpaqueMarker(n) {
			opaqueDirs = append(opaqueDirs, path.Dir(string(n.RealPath)))
			return nil
		}
		_, _, err := tree.AddPath(string(n.RealPath), newFileInfo(img, n.Reference))
		return err
	})
	if err != nil {
		log.WithFields("error", err).Debug("unable to build UI tree")
	}
	return tree, opaqueDirs
}
