package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/components"
	"github.com/wagoodman/dive/runtime/ui/extension_components"
	"github.com/wagoodman/dive/runtime/ui/extension_viewmodels"
	"github.com/wagoodman/dive/runtime/ui/viewmodels"
	"sync"
)

const debug = false

// type global
var (
	once         sync.Once
	appSingleton         *diveApp
)

type diveApp struct {
	app         *tview.Application
	layers       tview.Primitive
	fileTree       tview.Primitive
	filterView     tview.Primitive
}



func newApp(app *tview.Application, analysis *image.AnalysisResult, cache filetree.Comparer, isCNB bool) (*diveApp, error) {
	var err error
	once.Do(func() {
		//initialize viewmodels
		filterViewModel := viewmodels.NewFilterViewModel(nil)
		var layerModel viewmodels.LayersModel
		var layerDetailsView tview.Primitive
		if isCNB{
			cnbLayerViewModel := extension_viewmodels.NewCNBLayersViewModel(analysis.Layers, analysis.BOMMapping)
			cnbLayerDetailsView := extension_components.NewCNBLayerDetailsView(cnbLayerViewModel).Setup()
			layerModel = cnbLayerViewModel
			layerDetailsView = cnbLayerDetailsView
		} else {
			layerViewModel := viewmodels.NewLayersViewModel(analysis.Layers)
			regularLayerDetailsView := components.NewLayerDetailsView(layerViewModel).Setup()
			layerModel = layerViewModel
			layerDetailsView = regularLayerDetailsView
		}
		//layerViewModel := viewmodels.NewLayersViewModel(analysis.Layers)
		treeViewModel, err := viewmodels.NewTreeViewModel(cache, layerModel, filterViewModel)
		if err != nil {
			panic(err)
		}

		// initialize views
		imageDetails := components.NewImageDetailsView(analysis)

		filterView := components.NewFilterView(filterViewModel).Setup()
		layersView := components.NewLayerList(treeViewModel).Setup()
		fileTreeView := components.NewTreeView(treeViewModel).Setup()


		grid := tview.NewGrid()
		grid.SetRows(-4,-1,-1,1).SetColumns(-1,-1, 3)
		grid.SetBorder(false)
		grid.AddItem(layersView, 0,0,1,1,5, 10, true).
			AddItem(layerDetailsView,1,0,1,1,10,40, false).
			AddItem(imageDetails,2,0,1, 1,10,10,false).
			AddItem(fileTreeView, 0, 1, 3, 1, 0,0, true).
			AddItem(filterView, 3,0,1,2,0,0,false)


		switchFocus := func(event *tcell.EventKey) *tcell.EventKey {
			var result *tcell.EventKey = nil
			switch event.Key() {
			case tcell.KeyTAB:
				if appSingleton.app.GetFocus() == appSingleton.layers {
					appSingleton.app.SetFocus(appSingleton.fileTree)
				} else {
					appSingleton.app.SetFocus(appSingleton.layers)
				}
			case tcell.KeyCtrlF:
				if filterView.HasFocus() {
					filterView.Blur()
					appSingleton.app.SetFocus(grid)
				} else {
					appSingleton.app.SetFocus(filterView)
				}

			default:
				result = event
			}
			return result
		}

		grid.SetInputCapture(switchFocus)

		app.SetRoot(grid,true)
		appSingleton = &diveApp{
			app: app,
			fileTree: fileTreeView,
			layers: layersView,
		}
		app.SetFocus(layersView)
	})

	return appSingleton, err
}

// Run is the UI entrypoint.
func Run(analysis *image.AnalysisResult, treeStack filetree.Comparer, isCNB bool) error {

	app := tview.NewApplication()
	_, err := newApp(app, analysis, treeStack, isCNB)
	if err != nil {
		return err
	}

	if err = app.Run(); err != nil {
		return err
	}
	return nil
}