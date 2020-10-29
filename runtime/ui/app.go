package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/components"
	"github.com/wagoodman/dive/runtime/ui/viewmodels"
	"os"
	"path/filepath"
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
	layers       *components.LayerList
	fileTree       *components.TreeView
}



func newApp(app *tview.Application, analysis *image.AnalysisResult, cache filetree.Comparer) (*diveApp, error) {
	var err error
	once.Do(func() {
		//initialize viewmodels
		filterViewModel := viewmodels.NewFilterViewModel(nil)
		layerViewModel := viewmodels.NewLayersViewModel(analysis.Layers)
		treeViewModel, err := viewmodels.NewTreeViewModel(cache, layerViewModel, filterViewModel)
		if err != nil {
			panic(err)
		}

		// initialize views
		imageDetails := components.NewImageDetailsView(analysis)

		filterView := components.NewFilterView(treeViewModel).Setup()
		layerDetailsView := components.NewLayerDetailsView(treeViewModel).Setup()
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
				if appSingleton.layers.HasFocus() {
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
func Run(analysis *image.AnalysisResult, treeStack filetree.Comparer) error {
	debugFile := filepath.Join("/tmp", "dive","debug.out")
	LogOutputFile, _ := os.OpenFile(debugFile, os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0666)
	defer LogOutputFile.Close()
	logrus.SetOutput(LogOutputFile)
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Debugln("debug start:")
	app := tview.NewApplication()
	_, err := newApp(app, analysis, treeStack)
	if err != nil {
		return err
	}

	if err = app.Run(); err != nil {
		return err
	}
	return nil
}

func intMax(int1 ,int2 int) int {
	if int1 > int2 {
		return int1
	}
	return int2
}
