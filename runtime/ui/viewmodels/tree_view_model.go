package viewmodels

import (
	"fmt"
	"regexp"

	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/internal/log"
)

//go:generate faux --interface FilterModel --output fakes/fake_filter_model.go
type FilterModel interface {
	SetFilter(r *regexp.Regexp)
	GetFilter() *regexp.Regexp
}

//go:generate faux --interface LayersModel --output fakes/fake_layers_model.go
type LayersModel interface {
	SetLayerIndex(index int) bool
	GetCompareIndicies() filetree.TreeIndexKey
	GetCurrentLayer() *image.Layer
	GetPrintableLayers() []fmt.Stringer
	GetMode() LayerCompareMode
	SwitchLayerMode() error
}

//go:generate faux --interface TreeCache --output fakes/fake_tree_cache.go
type TreeCache interface {
	GetTree(key filetree.TreeIndexKey) (TreeModel, error)
}

//go:generate faux --interface TreeModel --output fakes/fake_tree_model.go
type TreeModel interface {
	StringBetween(start, stop int, showAttributes bool) string
	VisitDepthParentFirst(visitor filetree.Visitor, evaluator filetree.VisitEvaluator) error
	VisitDepthChildFirst(visitor filetree.Visitor, evaluator filetree.VisitEvaluator) error
	RemovePath(path string) error
	VisibleSize() int
	GetNode(path string) (*filetree.FileNode, error)
}

type TreeViewModel struct {
	currentTree     TreeModel
	cache           TreeCache
	hiddenDiffTypes []bool
	// Make this an interface that is composed with the FilterView
	FilterModel
	LayersModel
}

func NewTreeViewModel(cache TreeCache, lModel LayersModel, fModel FilterModel) (*TreeViewModel, error) {
	curTreeIndex := filetree.NewTreeIndexKey(0, 0, 0, 0)
	tree, err := cache.GetTree(curTreeIndex)
	if err != nil {
		return nil, err
	}
	return &TreeViewModel{
		currentTree:     tree,
		cache:           cache,
		hiddenDiffTypes: make([]bool, 4),
		FilterModel:     fModel,
		LayersModel:     lModel,
	}, nil
}

func (tvm *TreeViewModel) StringBetween(startRow, stopRow int, showAttributes bool) string {
	return tvm.currentTree.StringBetween(startRow, stopRow, showAttributes)
}

func (tvm *TreeViewModel) VisitDepthParentFirst(visitor filetree.Visitor, evaluator filetree.VisitEvaluator) error {
	return tvm.currentTree.VisitDepthParentFirst(visitor, evaluator)
}
func (tvm *TreeViewModel) VisitDepthChildFirst(visitor filetree.Visitor, evaluator filetree.VisitEvaluator) error {
	return tvm.currentTree.VisitDepthChildFirst(visitor, evaluator)

}
func (tvm *TreeViewModel) RemovePath(path string) error {
	return tvm.currentTree.RemovePath(path)
}
func (tvm *TreeViewModel) VisibleSize() int {
	return tvm.currentTree.VisibleSize()
}

func (tvm *TreeViewModel) GetNode(path string) (*filetree.FileNode, error) {
	return tvm.currentTree.GetNode(path)
}

func (tvm *TreeViewModel) SetFilter(filterRegex *regexp.Regexp) {
	tvm.FilterModel.SetFilter(filterRegex)
	if err := tvm.filterUpdate(); err != nil {
		// TODO -Dan- handle panics
		panic(err)
	}
}

// TODO: this seems like a very expensive operration, look for ways to optimize.
// TODO make type int a strongly typed argument
// TODO: handle errors correctly
func (tvm *TreeViewModel) ToggleHiddenFileType(filetype filetree.DiffType) bool {
	tvm.hiddenDiffTypes[filetype] = !tvm.hiddenDiffTypes[filetype]
	if err := tvm.filterUpdate(); err != nil {
		//panic(err)
		return false
	}
	return true
}

func (tvm *TreeViewModel) GetHiddenFileType(filetype filetree.DiffType) bool {
	return tvm.hiddenDiffTypes[filetype]
}

func (tvm *TreeViewModel) filterUpdate() error {
	log.WithFields(
		"model", "treeView",
	).Trace("updating filter")

	// keep the t selection in parity with the current DiffType selection
	filter := tvm.GetFilter()
	err := tvm.currentTree.VisitDepthChildFirst(func(node *filetree.FileNode) error {
		node.Data.ViewInfo.Hidden = tvm.hiddenDiffTypes[node.Data.DiffType]

		for _, child := range node.Children {
			if !child.Data.ViewInfo.Hidden {
				node.Data.ViewInfo.Hidden = false
				return nil
			}
		}

		if len(node.Children) > 0 {
			node.Data.ViewInfo.Hidden = true
		}

		if filter != nil && !node.Data.ViewInfo.Hidden { // hide nodes that do not match the current file filter regex (also don't unhide nodes that are already hidden)
			match := filter.FindString(node.Path())
			node.Data.ViewInfo.Hidden = len(match) == 0
		}
		return nil
	}, nil)

	if err != nil {
		log.Errorf("error updating filter on current tree: %s", err)
		return err
	}

	return nil
}

// Override functions
func (tvm *TreeViewModel) SetLayerIndex(index int) bool {
	var err error
	if tvm.LayersModel.SetLayerIndex(index) {
		err = tvm.setCurrentTree(tvm.GetCompareIndicies())
	}
	return err == nil
}

func (tvm *TreeViewModel) setCurrentTree(key filetree.TreeIndexKey) error {
	newTree, err := tvm.cache.GetTree(key)
	if err != nil {
		return err
	}

	// we should visit each node in the currently displayed tree and ensure that the next tree
	// has all of the current view state for all paths that exist in the new tree. This ensures that if a
	// user collapsed a directory, then that directory will stay collapsed when changing trees. This also
	// ensures that if a user changes back to the same tree again that the latest state is used from the
	// current tree selected and not from stale state from a previous user selection.
	tvm.currentTree.VisitDepthParentFirst(func(node *filetree.FileNode) error {
		newNode, err := newTree.GetNode(node.Path())
		if err == nil {
			newNode.Data.ViewInfo = node.Data.ViewInfo
		}
		return nil
	}, nil)

	tvm.currentTree = newTree
	if err := tvm.filterUpdate(); err != nil {
		return err
	}
	return nil
}

func (tvm *TreeViewModel) SwitchLayerMode() error {
	if err := tvm.LayersModel.SwitchLayerMode(); err != nil {
		return err
	}
	// TODO: Handle this error
	if err := tvm.setCurrentTree(tvm.GetCompareIndicies()); err != nil {
		return err
	}
	return nil
}
