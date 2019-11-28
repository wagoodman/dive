package view

import (
	"github.com/jroimartin/gocui"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
)

type Views struct {
	Tree    *FileTree
	Layer   *Layer
	Status  *Status
	Filter  *Filter
	Details *Details
	Debug   *Debug
}

func NewViews(g *gocui.Gui, analysis *image.AnalysisResult, cache filetree.Comparer) (*Views, error) {
	Layer, err := newLayerView(g, analysis.Layers)
	if err != nil {
		return nil, err
	}

	treeStack := analysis.RefTrees[0]
	Tree, err := newFileTreeView(g, treeStack, analysis.RefTrees, cache)
	if err != nil {
		return nil, err
	}

	Status := newStatusView(g)

	// set the layer view as the first selected view
	Status.SetCurrentView(Layer)

	Filter := newFilterView(g)

	Details := newDetailsView(g, analysis.Efficiency, analysis.Inefficiencies, analysis.SizeBytes)

	Debug := newDebugView(g)

	return &Views{
		Tree:    Tree,
		Layer:   Layer,
		Status:  Status,
		Filter:  Filter,
		Details: Details,
		Debug:   Debug,
	}, nil
}

func (views *Views) All() []Renderer {
	return []Renderer{
		views.Tree,
		views.Layer,
		views.Status,
		views.Filter,
		views.Details,
	}
}
