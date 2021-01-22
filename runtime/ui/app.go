package ui

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/components"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/viewmodels"
)

// type global
var (
	once        sync.Once
	uiSingleton *UI
)

type UI struct {
	app      *components.DiveApplication
	layers   tview.Primitive
	fileTree tview.Primitive
}

func newApp(app *tview.Application, analysis *image.AnalysisResult, cache filetree.Comparer) (*UI, error) {
	var err error
	once.Do(func() {

		// TODO: Extract initilaization logic into its own package
		format.SyncWithTermColors()

		config := components.NewKeyConfig()
		appConfig := components.AppConfig{}
		diveApplication := components.NewDiveApplication(app)

		//initialize viewmodels
		filterViewModel := viewmodels.NewFilterViewModel(nil)

		layerModel := viewmodels.NewLayersViewModel(analysis.Layers)
		regularLayerDetailsView := components.NewLayerDetailsView(layerModel).Setup()
		layerDetailsBox := components.NewWrapper("Layer Details", "", regularLayerDetailsView).Setup()
		layerDetailsBox.SetVisibility(components.MinHeightVisibility(10))

		//layerViewModel := viewmodels.NewLayersViewModel(analysis.Layers)
		cacheWrapper := CacheWrapper{Cache: &cache}
		treeViewModel, err := viewmodels.NewTreeViewModel(&cacheWrapper, layerModel, filterViewModel)
		if err != nil {
			panic(err)
		}

		// initialize views
		imageDetailsView := components.NewImageDetailsView(analysis).Setup()
		imageDetailsBox := components.NewWrapper("Image Details", "", imageDetailsView).Setup()
		imageDetailsBox.SetVisibility(components.MinHeightVisibility(10))

		filterView := components.NewFilterView(treeViewModel).Setup()

		layersView := components.NewLayerList(treeViewModel).Setup(config)

		layerSubtitle := fmt.Sprintf("Cmp%7s  %s", "Size", "Command")
		layersBox := components.NewWrapper("Layers", layerSubtitle, layersView).Setup()

		fileTreeView := components.NewTreeView(treeViewModel)
		fileTreeView = fileTreeView.Setup(config)
		fileTreeBox := components.NewWrapper("Current Layer Contents", "", fileTreeView).Setup()

		keyMenuView := components.NewKeyMenuView()

		leftVisibleGrid := components.NewVisibleFlex()
		leftVisibleGrid.SetDirection(tview.FlexRow)
		rightVisibleGrid := components.NewVisibleFlex()
		rightVisibleGrid.SetDirection(tview.FlexRow)
		totalVisibleGrid := components.NewVisibleFlex()
		gridWithFooter := tview.NewGrid().
			SetRows(0, 1).
			SetColumns(0).
			AddItem(totalVisibleGrid, 0, 0, 1, 1, 0, 0, true).
			AddItem(keyMenuView, 1, 0, 1, 1, 0, 0, false)

		leftVisibleGrid.AddItem(layersBox, 0, 3, true).
			AddItem(layerDetailsBox, 0, 1, false).
			AddItem(imageDetailsBox, 0, 1, false).
			SetConsumers(layerDetailsBox, layersBox).
			SetConsumers(imageDetailsBox, layersBox)

		rightVisibleGrid.AddItem(fileTreeBox, 0, 1, false).
			AddItem(filterView, 1, 0, false).
			SetConsumers(filterView, fileTreeBox)

		leftPortion, rightPortion := appConfig.GetPaneWidth()
		totalVisibleGrid.AddItem(leftVisibleGrid, 0, leftPortion, true).
			AddItem(rightVisibleGrid, 0, rightPortion, false)

		uiSingleton = &UI{
			app:      diveApplication,
			fileTree: fileTreeBox,
			layers:   layersBox,
		}

		keyMenuView.AddBoundViews(diveApplication)

		quitBinding, err := config.GetKeyBinding("keybinding.quit")
		if err != nil {
			// TODO handle this as an error
			panic(err)
		}

		filterBinding, err := config.GetKeyBinding("keybinding.filter-files")
		if err != nil {
			// TODO handle this as an error
			panic(err)
		}
		switchBinding, err := config.GetKeyBinding("keybinding.toggle-view")
		if err != nil {
			// TODO handle this as an error
			panic(err)
		}
		diveApplication.AddBindings(quitBinding, filterBinding, switchBinding)
		diveApplication.AddBoundViews(fileTreeBox, layersBox, filterView)

		switchFocus := func(event *tcell.EventKey) *tcell.EventKey {
			var result *tcell.EventKey = nil
			switch {
			case quitBinding.Match(event):
				app.Stop()
			case switchBinding.Match(event):
				if diveApplication.GetFocus() == uiSingleton.layers {
					diveApplication.SetFocus(uiSingleton.fileTree)
				} else {
					diveApplication.SetFocus(uiSingleton.layers)
				}
			case filterBinding.Match(event):
				if filterView.HasFocus() {
					filterView.Blur()
					diveApplication.SetFocus(fileTreeBox)
				} else {
					diveApplication.SetFocus(filterView)
				}

			default:
				result = event
			}
			return result
		}

		diveApplication.SetInputCapture(switchFocus)

		diveApplication.SetRoot(gridWithFooter, true)
		diveApplication.SetFocus(gridWithFooter)

		// additional setup configuration
		if appConfig.GetAggregateLayerSetting() {
			err := layerModel.SwitchLayerMode()
			if err != nil {
				panic(err)
			}
		}

		if appConfig.GetCollapseDir() {
			fileTreeView.CollapseOrExpandAll()
		}

		for _, hideType := range appConfig.GetDefaultHide() {
			treeViewModel.ToggleHiddenFileType(hideType)
		}

		if appConfig.GetShowAttributes() {
			fileTreeView.ToggleHideAttributes()
		}
	})

	return uiSingleton, err
}

// Run is the UI entrypoint.
func Run(analysis *image.AnalysisResult, treeStack filetree.Comparer) error {
	app := tview.NewApplication()
	_, err := newApp(app, analysis, treeStack)
	if err != nil {
		return err
	}

	if err = uiSingleton.app.Run(); err != nil {
		logrus.Error("app error: ", err.Error())
		return err
	}
	logrus.Info("app run loop exited")
	return nil
}

// TODO move me to initialization package
type CacheWrapper struct {
	Cache *filetree.Comparer
}

func (c *CacheWrapper) GetTree(key filetree.TreeIndexKey) (viewmodels.TreeModel, error) {
	return c.Cache.GetTree(key)
}
