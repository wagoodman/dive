package ui

import (
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/components"
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
		config := components.NewKeyConfig()

		// ensure the background color is inherited from the terminal emulator
		//tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
		//tview.Styles.PrimaryTextColor = tcell.ColorDefault

		//initialize viewmodels
		filterViewModel := viewmodels.NewFilterViewModel(nil)
		var layerModel viewmodels.LayersModel
		var layerDetailsBox *components.Wrapper
		if isCNB {
			cnbLayerViewModel := viewmodels.NewCNBLayersViewModel(analysis.Layers, analysis.BOMMapping)
			cnbLayerDetailsView := components.NewCNBLayerDetailsView(cnbLayerViewModel).Setup()
			layerModel = cnbLayerViewModel
			layerDetailsBox = components.NewWrapper("CNB Layer Details", "", cnbLayerDetailsView).Setup()
		} else {
			layerViewModel := viewmodels.NewLayersViewModel(analysis.Layers)
			regularLayerDetailsView := components.NewLayerDetailsView(layerViewModel).Setup()
			layerModel = layerViewModel
			layerDetailsBox = components.NewWrapper("Layer Details", "", regularLayerDetailsView).Setup()
		}
		layerDetailsBox.SetVisibility(components.MinHeightVisibility(10))

		//layerViewModel := viewmodels.NewLayersViewModel(analysis.Layers)
		treeViewModel, err := viewmodels.NewTreeViewModel(cache, layerModel, filterViewModel)
		if err != nil {
			// TODO: replace panic with a reasonable exit strategy
			panic(err)
		}

		// initialize views
		imageDetailsView := components.NewImageDetailsView(analysis)
		imageDetailsBox := components.NewWrapper("Image Details", "", imageDetailsView).Setup()
		imageDetailsBox.SetVisibility(components.MinHeightVisibility(10))

		filterView := components.NewFilterView(treeViewModel).Setup()

		layersView := components.NewLayerList(treeViewModel).Setup(config)
		layersBox := components.NewWrapper("Layers", "subtitle!", layersView).Setup()

		fileTreeView := components.NewTreeView(treeViewModel)
		fileTreeView = fileTreeView.Setup(config)
		fileTreeBox := components.NewWrapper("Current Layer Contents", "subtitle!", fileTreeView).Setup()

		// Implementation notes: should we factor out this setup??
		// Probably yes, but difficult to make this both easy to setup & mutable
		leftVisibleGrid := components.NewVisibleFlex()
		leftVisibleGrid.SetDirection(tview.FlexRow)
		rightVisibleGrid := components.NewVisibleFlex()
		rightVisibleGrid.SetDirection(tview.FlexRow)
		totalVisibleGrid := components.NewVisibleFlex()

		leftVisibleGrid.AddItem(layersBox, 0, 3, true).
			AddItem(layerDetailsBox, 0, 1, false).
			AddItem(imageDetailsBox, 0, 1, false).
			SetConsumers(layerDetailsBox, layersBox).
			SetConsumers(imageDetailsBox, layersBox)

		rightVisibleGrid.AddItem(fileTreeBox, 0, 1, false).
			AddItem(filterView, 1, 0, false).
			SetConsumers(filterView, fileTreeBox)

		totalVisibleGrid.AddItem(leftVisibleGrid, 0, 1, true).
			AddItem(rightVisibleGrid, 0, 1, false)

		appSingleton = &diveApp{
			app:      app,
			fileTree: fileTreeBox,
			layers:   layersBox,
		}

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
					appSingleton.app.SetFocus(fileTreeBox)
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
		app.SetFocus(totalVisibleGrid)
	})

	return appSingleton, err
}

// Run is the UI entrypoint.
func Run(analysis *image.AnalysisResult, treeStack filetree.Comparer, isCNB bool) error {
	cfg := zap.NewDevelopmentConfig()
	os.MkdirAll("/tmp/dive", os.ModePerm)
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

	if err = appSingleton.app.Run(); err != nil {
		zap.S().Info("app error: ", err.Error())
		return err
	}
	zap.S().Info("app run loop exited")
	return nil
}
