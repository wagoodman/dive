package ui

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/utils"
	"github.com/wagoodman/keybinding"
	"log"

	"github.com/jroimartin/gocui"
	"github.com/lunixbochs/vtclean"
	"github.com/wagoodman/dive/image"
	"strings"
)

// LayerView holds the UI objects and data models for populating the lower-left pane. Specifically the pane that
// shows the image layers and layer selector.
type LayerView struct {
	Name              string
	gui               *gocui.Gui
	view              *gocui.View
	header            *gocui.View
	LayerIndex        int
	Layers            []image.Layer
	CompareMode       CompareType
	CompareStartIndex int
	ImageSize         uint64

	keybindingCompareAll   []keybinding.Key
	keybindingCompareLayer []keybinding.Key
}

// NewDetailsView creates a new view object attached the the global [gocui] screen object.
func NewLayerView(name string, gui *gocui.Gui, layers []image.Layer) (layerView *LayerView) {
	layerView = new(LayerView)

	// populate main fields
	layerView.Name = name
	layerView.gui = gui
	layerView.Layers = layers

	switch mode := viper.GetBool("layer.show-aggregated-changes"); mode {
	case true:
		layerView.CompareMode = CompareAll
	case false:
		layerView.CompareMode = CompareLayer
	default:
		utils.PrintAndExit(fmt.Sprintf("unknown layer.show-aggregated-changes value: %v", mode))
	}

	var err error
	layerView.keybindingCompareAll, err = keybinding.ParseAll(viper.GetString("keybinding.compare-all"))
	if err != nil {
		log.Panicln(err)
	}

	layerView.keybindingCompareLayer, err = keybinding.ParseAll(viper.GetString("keybinding.compare-layer"))
	if err != nil {
		log.Panicln(err)
	}

	return layerView
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (view *LayerView) Setup(v *gocui.View, header *gocui.View) error {

	// set view options
	view.view = v
	view.view.Editable = false
	view.view.Wrap = false
	view.view.Frame = false

	view.header = header
	view.header.Editable = false
	view.header.Wrap = false
	view.header.Frame = false

	// set keybindings
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowDown, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorDown() }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowUp, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorUp() }); err != nil {
		return err
	}

	for _, key := range view.keybindingCompareLayer {
		if err := view.gui.SetKeybinding(view.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return view.setCompareMode(CompareLayer) }); err != nil {
			return err
		}
	}

	for _, key := range view.keybindingCompareAll {
		if err := view.gui.SetKeybinding(view.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return view.setCompareMode(CompareAll) }); err != nil {
			return err
		}
	}

	return view.Render()
}

// IsVisible indicates if the layer view pane is currently initialized.
func (view *LayerView) IsVisible() bool {
	if view == nil {
		return false
	}
	return true
}

// CursorDown moves the cursor down in the layer pane (selecting a higher layer).
func (view *LayerView) CursorDown() error {
	if view.LayerIndex < len(view.Layers) {
		err := CursorDown(view.gui, view.view)
		if err == nil {
			view.SetCursor(view.LayerIndex + 1)
		}
	}
	return nil
}

// CursorUp moves the cursor up in the layer pane (selecting a lower layer).
func (view *LayerView) CursorUp() error {
	if view.LayerIndex > 0 {
		err := CursorUp(view.gui, view.view)
		if err == nil {
			view.SetCursor(view.LayerIndex - 1)
		}
	}
	return nil
}

// SetCursor resets the cursor and orients the file tree view based on the given layer index.
func (view *LayerView) SetCursor(layer int) error {
	view.LayerIndex = layer
	Views.Tree.setTreeByLayer(view.getCompareIndexes())
	Views.Details.Render()
	view.Render()

	return nil
}

// currentLayer returns the Layer object currently selected.
func (view *LayerView) currentLayer() image.Layer {
	return view.Layers[(len(view.Layers)-1)-view.LayerIndex]
}

// setCompareMode switches the layer comparison between a single-layer comparison to an aggregated comparison.
func (view *LayerView) setCompareMode(compareMode CompareType) error {
	view.CompareMode = compareMode
	Update()
	Render()
	return Views.Tree.setTreeByLayer(view.getCompareIndexes())
}

// getCompareIndexes determines the layer boundaries to use for comparison (based on the current compare mode)
func (view *LayerView) getCompareIndexes() (bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) {
	bottomTreeStart = view.CompareStartIndex
	topTreeStop = view.LayerIndex

	if view.LayerIndex == view.CompareStartIndex {
		bottomTreeStop = view.LayerIndex
		topTreeStart = view.LayerIndex
	} else if view.CompareMode == CompareLayer {
		bottomTreeStop = view.LayerIndex - 1
		topTreeStart = view.LayerIndex
	} else {
		bottomTreeStop = view.CompareStartIndex
		topTreeStart = view.CompareStartIndex + 1
	}

	return bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop
}

// renderCompareBar returns the formatted string for the given layer.
func (view *LayerView) renderCompareBar(layerIdx int) string {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop := view.getCompareIndexes()
	result := "  "

	if layerIdx >= bottomTreeStart && layerIdx <= bottomTreeStop {
		result = Formatting.CompareBottom("  ")
	}
	if layerIdx >= topTreeStart && layerIdx <= topTreeStop {
		result = Formatting.CompareTop("  ")
	}

	return result
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (view *LayerView) Update() error {
	view.ImageSize = 0
	for idx := 0; idx < len(view.Layers); idx++ {
		view.ImageSize += view.Layers[idx].Size()
	}
	return nil
}

// Render flushes the state objects to the screen. The layers pane reports:
// 1. the layers of the image + metadata
// 2. the current selected image
func (view *LayerView) Render() error {

	// indicate when selected
	title := "Layers"
	if view.gui.CurrentView() == view.view {
		title = "● " + title
	}

	view.gui.Update(func(g *gocui.Gui) error {
		// update header
		view.header.Clear()
		width, _ := g.Size()
		headerStr := fmt.Sprintf("[%s]%s\n", title, strings.Repeat("─", width*2))
		headerStr += fmt.Sprintf("Cmp "+image.LayerFormat, "Image ID", "Size", "Command")
		fmt.Fprintln(view.header, Formatting.Header(vtclean.Clean(headerStr, false)))

		// update contents
		view.view.Clear()
		for revIdx := len(view.Layers) - 1; revIdx >= 0; revIdx-- {
			layer := view.Layers[revIdx]
			idx := (len(view.Layers) - 1) - revIdx

			layerStr := layer.String()
			compareBar := view.renderCompareBar(idx)

			if idx == view.LayerIndex {
				fmt.Fprintln(view.view, compareBar+"  "+Formatting.Selected(layerStr))
			} else {
				fmt.Fprintln(view.view, compareBar+"  "+layerStr)
			}

		}
		return nil
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected.
func (view *LayerView) KeyHelp() string {
	return renderStatusOption(view.keybindingCompareLayer[0].String(), "Show layer changes", view.CompareMode == CompareLayer) +
		renderStatusOption(view.keybindingCompareAll[0].String(), "Show aggregated changes", view.CompareMode == CompareAll)
}
