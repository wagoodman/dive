package ui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/components"
	"os"
	"path/filepath"
	"regexp"
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

//type Cache interface {
//	GetTree(key filetree.TreeIndexKey) (*filetree.FileTree, error)
//}


//func updateFileTree()
//
//func NewLayerListHandler(cache filetree.Comparer, analysis image.AnalysisResult,layerDetails tview.TextView) components.LayerListHandler {
//	return func(i int, stringer fmt.Stringer, r rune) {
//		bottomStart := intMax(0,i-1) // no values less than zero
//		bottomStop := intMax(0, i-1)
//		curTreeIndex := filetree.NewTreeIndexKey(bottomStart,bottomStop,i,i)
//		curTree, err := cache.GetTree(curTreeIndex)
//		layerDetails.SetText(components.LayerDetailsText(analysis.Layers[i]))
//		if err != nil {
//			panic(err)
//		}
//
//		fileTreeView.SetTree(curTree)
//	}
//}


func newApp(app *tview.Application, analysis *image.AnalysisResult, cache filetree.Comparer) (*diveApp, error) {
	var err error
	once.Do(func() {

		layersView := components.NewLayerList(nil)
		layersView.SetSubtitle("Cmp   Size  Command").SetBorder(true).SetTitle("Layers")

		curTreeIndex := filetree.NewTreeIndexKey(0,0,0,0)
		curTree, err := cache.GetTree(curTreeIndex)
		if err != nil {
			panic(err)
		}

		fileTreeView := components.NewTreeView(curTree)
		fileTreeView.SetTitle("Files").SetBorder(true)

		layerDetails := tview.NewTextView()
		layerDetails.SetTitle("Layer Details")
		layerDetails.SetDynamicColors(true).SetBorder(true)
		layerDetails.SetText(components.LayerDetailsText(analysis.Layers[0]))

		for _, layer := range analysis.Layers {
			layersView.AddItem(layer)
		}
		layersView.SetChangedFunc(func(i int, stringer fmt.Stringer, r rune) {
			bottomStart := intMax(0,i-1) // no values less than zero
			bottomStop := intMax(0, i-1)
			curTreeIndex := filetree.NewTreeIndexKey(bottomStart,bottomStop,i,i)
			curTree, err = cache.GetTree(curTreeIndex)
			layerDetailText := components.LayerDetailsText(analysis.Layers[i])
			layerDetails.SetText(layerDetailText)
			if err != nil {
				panic(err)
			}

			fileTreeView.SetTree(curTree)
		})

		imageDetails := components.NewImageDetailsView(analysis)
		grid := tview.NewGrid()
		filterView := components.NewFilterView()
		filterView.SetChangedFunc(
			func(textToCheck string) {
				var filterRegex *regexp.Regexp = nil
				var err error

				if len(textToCheck) > 0 {
					filterRegex, err = regexp.Compile(textToCheck)
					if err != nil {
						return
					}
				}

				fileTreeView.SetFilterRegex(filterRegex)
				return
			}).SetDoneFunc(func(key tcell.Key) {
				switch {
				case key == tcell.KeyEnter:
				app.SetFocus(grid)
				}
			})

		grid.SetRows(-4,-1,-1,1).SetColumns(-1,-1, 3)
		grid.SetBorder(false)
		grid.AddItem(layersView, 0,0,1,1,5, 10, true).
			AddItem(layerDetails,1,0,1,1,10,40, false).
			AddItem(imageDetails,2,0,1, 1,10,10,false).
			AddItem(fileTreeView, 0, 1, 3, 1, 0,0, true).
			AddItem(filterView, 3,0,1,2,0,0,false)


		switchFocus := func(event *tcell.EventKey) *tcell.EventKey {
			var result *tcell.EventKey = nil
			switch event.Key() {
			case tcell.KeyTAB:
				//fmt.Println("Tab")
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
		app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			logrus.Debugf("application handling in put %s\n", event.Name())
			return event
		})
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
