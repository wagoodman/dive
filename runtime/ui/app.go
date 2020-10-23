package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/components"
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
	finderFocus tview.Primitive
}


func newApp(app *tview.Application, analysis *image.AnalysisResult, cache filetree.Comparer) (*diveApp, error) {
	var err error
	once.Do(func() {

		layersView := components.NewLayerList([]string{})
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
			layersView.AddItem(layer.String()).SetChangedFunc(func(i int, s string, r rune) {
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
		}

		imageDetails := components.NewImageDetailsView(analysis)

		grid := tview.NewGrid().SetRows(-4,-1,-1)
		grid.SetBorder(false)
		grid.AddItem(layersView, 0,0,1,1,5, 10, false).
			AddItem(layerDetails,1,0,1,1,10,10, false).
			AddItem(imageDetails,2,0,1, 1,10,10,false)



		flex := tview.NewFlex().
			AddItem(grid, 0, 1, true).
			AddItem(fileTreeView, 0, 1, false)

		switchFocus := func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyTAB:
				if appSingleton.layers.HasFocus() {
					appSingleton.app.SetFocus(appSingleton.fileTree)
				} else {
					appSingleton.app.SetFocus(appSingleton.layers)
				}
				return nil
			default:
				return event
			}
		}

		app.SetInputCapture(switchFocus)

		app.SetRoot(flex,true).SetFocus(layersView)
		appSingleton = &diveApp{
			app: app,
			fileTree: fileTreeView,
			layers: layersView,
		}
	})

	once.Do(func() {
		curTreeIndex := filetree.NewTreeIndexKey(0,0,0,0)
		curTree, err := cache.GetTree(curTreeIndex)
		if err != nil {
			panic(err)
		}
		fileTreeView := components.NewTreeView(curTree)
		fileTreeView.SetTitle("Files").SetBorder(true)
		app.SetRoot(fileTreeView, true).SetFocus(fileTreeView)
		appSingleton = &diveApp{
			app: app,
			fileTree: fileTreeView,
			layers: nil,
		}
	})
	return appSingleton, err
}

// Run is the UI entrypoint.
func Run(analysis *image.AnalysisResult, treeStack filetree.Comparer) error {
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
