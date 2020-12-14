package ui

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/components"
	"github.com/wagoodman/dive/runtime/ui/extension_components"
	"github.com/wagoodman/dive/runtime/ui/extension_viewmodels"
	"github.com/wagoodman/dive/runtime/ui/viewmodels"
	"go.uber.org/zap"
)

const debug = false

// type global
var (
	once         sync.Once
	appSingleton *diveApp
)

type diveApp struct {
	app        *tview.Application
	layers     tview.Primitive
	fileTree   tview.Primitive
	filterView tview.Primitive
}


func newApp(app *tview.Application, analysis *image.AnalysisResult, cache filetree.Comparer, isCNB bool) (*diveApp, error) {
	var err error
	once.Do(func() {
		// ensure the background color is inherited from the terminal emulator
		//tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
		//tview.Styles.PrimaryTextColor = tcell.ColorDefault

		//initialize viewmodels
		filterViewModel := viewmodels.NewFilterViewModel(nil)
		var layerModel viewmodels.LayersModel
		var layerDetailsView tview.Primitive
		if isCNB {
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
		imageDetailsView := components.NewImageDetailsView(analysis)
		imageDetailsBox := components.NewWrapper("Image Details", "", imageDetailsView).Setup()

		filterView := components.NewFilterView(filterViewModel).Setup()

		layersView := components.NewLayerList(treeViewModel).Setup()
		layersBox := components.NewWrapper("Layers", "subtitle!", layersView).Setup()

		fileTreeView := components.NewTreeView(treeViewModel).Setup()
		fileTreeBox := components.NewWrapper("Current Layer Contents", "subtitle!", fileTreeView).Setup()

		// Implementation notes: should we factor out this setup??
		leftVisibleGrid := components.NewVisibleFlex()
		leftVisibleGrid.SetDirection(tview.FlexRow)
		rightVisibleGrid := components.NewVisibleFlex()
		rightVisibleGrid.SetDirection(tview.FlexRow)
		totalVisibleGrid := components.NewVisibleFlex()

		//
		visibleLayersView := components.NewVisibleWrapper(layersView)
		visibleLayerDetails := components.NewVisibleWrapper(layerDetailsView)
		visibleImageDetails := components.NewVisibleWrapper(imageDetailsBox)
		visibleFilterView := components.NewVisibleWrapper(filterView)

		// this iterface needs some work we should NOT be using closures...
		visibleFilterView.SetVisibility(func(p tview.Primitive) bool {
			zap.S().Info("  -- visible filter is ", !filterView.Empty() || filterView.HasFocus())
			return !filterView.Empty() || filterView.HasFocus()
		})

		visibleFileTreeView := components.NewVisibleWrapper(fileTreeBox)

		leftVisibleGrid.AddItem(visibleLayersView, 0, 2, true).
			AddItem(visibleLayerDetails, 0, 1, false).
			AddItem(visibleImageDetails, 0, 1, false)

			// TODO: make sure we use BOX styling set up by wagoodman
		rightVisibleGrid.AddItem(visibleFileTreeView, 0, 1, false).
			AddItem(visibleFilterView, 1, 0, false).
			SetConsumers(visibleFilterView, []int{0})

		totalVisibleGrid.AddItem(leftVisibleGrid, 0, 1, true).
			AddItem(rightVisibleGrid, 0, 1, false)

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
					appSingleton.app.SetFocus(visibleFileTreeView)
				} else {
					appSingleton.app.SetFocus(filterView)
				}

			default:
				result = event
			}
			return result
		}

		totalVisibleGrid.SetInputCapture(switchFocus)

		app.SetRoot(totalVisibleGrid, true)
		appSingleton = &diveApp{
			app:      app,
			fileTree: fileTreeBox,
			layers:   layersBox,
		}
		app.SetFocus(layersBox)
	})

	return appSingleton, err
}

// Run is the UI entrypoint.
func Run(analysis *image.AnalysisResult, treeStack filetree.Comparer, isCNB bool) error {
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{"/tmp/dive/debug.out"}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	defer logger.Sync() // flushes buffer, if any
	logger.Sugar().Debug("Debug Start")

	zap.S().Info("Starting Hidden Flex Program")

	app := tview.NewApplication()
	_, err = newApp(app, analysis, treeStack, isCNB)
	if err != nil {
		return err
	}

	if err = app.Run(); err != nil {
		return err
	}
	return nil
}
