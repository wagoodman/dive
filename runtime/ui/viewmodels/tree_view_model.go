package viewmodels

import (
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
)

type FilterModel interface {
	SetFilter(r *regexp.Regexp)
	GetFilter() *regexp.Regexp
}

type LayersModel interface {
	SetLayerIndex(index int) bool
	GetCompareIndicies() filetree.TreeIndexKey
	GetCurrentLayer() *image.Layer
	GetPrintableLayers() []fmt.Stringer
	GetMode() LayerCompareMode
	SwitchMode()
}

type TreeViewModel struct {
	currentTree     *filetree.FileTree
	cache           filetree.Comparer
	hiddenDiffTypes []bool
	// Make this an interface that is composed with the FilterView
	FilterModel
	LayersModel
}

func NewTreeViewModel(cache filetree.Comparer, lModel LayersModel, fModel FilterModel) (*TreeViewModel, error) {
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

func (tvm *TreeViewModel) SetFilter(filterRegex *regexp.Regexp) {
	tvm.FilterModel.SetFilter(filterRegex)
	if err := tvm.FilterUpdate(); err != nil {
		panic(err)
	}
}

// TODO: this seems like a very expensive operration, look for ways to optimize.
// TODO make type int a strongly typed argument
// TODO: handle errors correctly
func (tvm *TreeViewModel) ToggleHiddenFileType(filetype filetree.DiffType) bool {
	tvm.hiddenDiffTypes[filetype] = !tvm.hiddenDiffTypes[filetype]
	if err := tvm.FilterUpdate(); err != nil {
		//panic(err)
		return false
	}
	return true

}

// TODO: maek this method private, cant think of a reason for this to be public
func (tvm *TreeViewModel) FilterUpdate() error {
	logrus.Debug("Updating filter!!!")
	// keep the t selection in parity with the current DiffType selection
	filter := tvm.GetFilter()
	err := tvm.currentTree.VisitDepthChildFirst(func(node *filetree.FileNode) error {
		node.Data.ViewInfo.Hidden = tvm.hiddenDiffTypes[node.Data.DiffType]

		visibleChild := false
		for _, child := range node.Children {
			if !child.Data.ViewInfo.Hidden {
				visibleChild = true
				node.Data.ViewInfo.Hidden = false
				return nil
			}
		}

		if filter != nil && !node.Data.ViewInfo.Hidden && !visibleChild { // hide nodes that do not match the current file filter regex (also don't unhide nodes that are already hidden)
			match := filter.FindString(node.Path())
			node.Data.ViewInfo.Hidden = len(match) == 0
		}
		return nil
	}, nil)

	if err != nil {
		logrus.Errorf("unable to propagate t model tree: %+v", err)
		return err
	}

	return nil
}

// Override functions
func (tvm *TreeViewModel) SetLayerIndex(index int) bool {
	if tvm.LayersModel.SetLayerIndex(index) {
		err := tvm.setCurrentTree(tvm.GetCompareIndicies())
		if err != nil {
			// TODO handle error here
			return false
		}
		return true
	}
	return false
}

func (tvm *TreeViewModel) setCurrentTree(key filetree.TreeIndexKey) error {
	collapsedList := map[string]interface{}{}

	newTree, err := tvm.cache.GetTree(key)
	if err != nil {
		return err
	}

	evaluateFunc := func(node *filetree.FileNode) bool {
		if node.Parent != nil && node.Parent.Data.ViewInfo.Hidden {
			return false
		}
		return true
	}

	tvm.currentTree.VisitDepthParentFirst(func(node *filetree.FileNode) error {
		if node.Data.ViewInfo.Collapsed {
			collapsedList[node.Path()] = true
		}
		return nil
	}, evaluateFunc)

	newTree.VisitDepthParentFirst(func(node *filetree.FileNode) error {
		_, ok := collapsedList[node.Path()]
		if ok {
			node.Data.ViewInfo.Collapsed = true
		}
		return nil
	}, evaluateFunc)

	tvm.currentTree = newTree
	if err := tvm.FilterUpdate(); err != nil {
		return err
	}
	return nil
}

func (tvm *TreeViewModel) SwitchMode() {
	tvm.LayersModel.SwitchMode()
	// TODO: Handle this error
	tvm.setCurrentTree(tvm.GetCompareIndicies())
}
