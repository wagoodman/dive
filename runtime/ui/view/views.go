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
	all     []*Renderer
}

func NewViews(g *gocui.Gui, analysis *image.AnalysisResult, cache filetree.Comparer) (*Views, error) {
	Layer, err := newLayerView("layers", g, analysis.Layers)
	if err != nil {
		return nil, err
	}

	treeStack := analysis.RefTrees[0]
	Tree, err := newFileTreeView("filetree", g, treeStack, analysis.RefTrees, cache)
	if err != nil {
		return nil, err
	}

	Status := newStatusView("status", g)

	// set the layer view as the first selected view
	Status.SetCurrentView(Layer)

	Filter := newFilterView("filter", g)

	Details := newDetailsView("details", g, analysis.Efficiency, analysis.Inefficiencies, analysis.SizeBytes)

	return &Views{
		Tree:    Tree,
		Layer:   Layer,
		Status:  Status,
		Filter:  Filter,
		Details: Details,
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
