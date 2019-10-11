package ui

import (
	"github.com/jroimartin/gocui"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
)

// var ccOnce sync.Once
var controllers *controllerCollection

type controllerCollection struct {
	Tree    *fileTreeController
	Layer   *layerController
	Status  *statusController
	Filter  *filterController
	Details *detailsController
	lookup  map[string]Controller
}

func newControllerCollection(g *gocui.Gui, analysis *image.AnalysisResult, cache filetree.TreeCache) (*controllerCollection, error) {
	var err error

	controllers = &controllerCollection{}
	controllers.lookup = make(map[string]Controller)

	controllers.Layer, err = newLayerController("layers", g, analysis.Layers)
	if err != nil {
		return nil, err
	}
	controllers.lookup[controllers.Layer.name] = controllers.Layer

	treeStack, err := filetree.StackTreeRange(analysis.RefTrees, 0, 0)
	if err != nil {
		return nil, err
	}
	controllers.Tree, err = newFileTreeController("filetree", g, treeStack, analysis.RefTrees, cache)
	if err != nil {
		return nil, err
	}
	controllers.lookup[controllers.Tree.name] = controllers.Tree

	controllers.Status = newStatusController("status", g)
	controllers.lookup[controllers.Status.name] = controllers.Status

	controllers.Filter = newFilterController("filter", g)
	controllers.lookup[controllers.Filter.name] = controllers.Filter

	controllers.Details = newDetailsController("details", g, analysis.Efficiency, analysis.Inefficiencies)
	controllers.lookup[controllers.Details.name] = controllers.Details
	return controllers, nil
}
