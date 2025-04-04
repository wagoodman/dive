package view

import (
	"github.com/awesome-gocui/gocui"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1"
)

type View interface {
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

func NewViews(g *gocui.Gui, cfg v1.Config) (*Views, error) {
	layer, err := newLayerView(g, cfg)
	if err != nil {
		return nil, err
	}

	tree, err := newFileTreeView(g, cfg, 0)
	if err != nil {
		return nil, err
	}

	status := newStatusView(g)

	// set the layer view as the first selected view
	status.SetCurrentView(layer)

	return &Views{
		Tree:   tree,
		Layer:  layer,
		Status: status,
		Filter: newFilterView(g),
		ImageDetails: &ImageDetails{
			gui:            g,
			imageName:      cfg.Analysis.Image,
			imageSize:      cfg.Analysis.SizeBytes,
			efficiency:     cfg.Analysis.Efficiency,
			inefficiencies: cfg.Analysis.Inefficiencies,
			kb:             cfg.Preferences.KeyBindings,
		},
		LayerDetails: &LayerDetails{gui: g, kb: cfg.Preferences.KeyBindings},
		Debug:        newDebugView(g),
	}, nil
}

func (views *Views) Renderers() []Renderer {
	return []Renderer{
		views.Tree,
		views.Layer,
		views.Status,
		views.Filter,
		views.LayerDetails,
		views.ImageDetails,
	}
}
