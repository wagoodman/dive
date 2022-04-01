package view

import (
	"github.com/awesome-gocui/gocui"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
)

type IView interface {
	Setup(*gocui.View, *gocui.View) error
	Name() string
	IsVisible() bool
}

type Views struct {
	Tree         *FileTree
	Layer        *Layer
	Status       *Status
	Filter       *Filter
	LayerDetails *LayerDetails
	ImageDetails *ImageDetails
	Debug        *Debug
}

var _ []IView = []IView{
	&FileTree{},
	&Layer{},
	&Filter{},
	&LayerDetails{},
	&ImageDetails{},
	&Debug{},
}

func NewViews(g *gocui.Gui, imageName string, analysis *image.AnalysisResult, cache filetree.Comparer) (*Views, error) {
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

	LayerDetails := &LayerDetails{gui: g}
	ImageDetails := &ImageDetails{
		gui:            g,
		imageName:      imageName,
		imageSize:      analysis.SizeBytes,
		efficiency:     analysis.Efficiency,
		inefficiencies: analysis.Inefficiencies,
	}

	Debug := newDebugView(g)

	return &Views{
		Tree:         Tree,
		Layer:        Layer,
		Status:       Status,
		Filter:       Filter,
		ImageDetails: ImageDetails,
		LayerDetails: LayerDetails,
		Debug:        Debug,
	}, nil
}

func (views *Views) All() []Renderer {
	return []Renderer{
		views.Tree,
		views.Layer,
		views.Status,
		views.Filter,
		views.LayerDetails,
		views.ImageDetails,
	}
}
