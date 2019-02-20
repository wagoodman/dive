package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/lunixbochs/vtclean"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/image"
	"github.com/wagoodman/dive/utils"
	"github.com/wagoodman/keybinding"
	"strings"
)

// LayerController holds the UI objects and data models for populating the lower-left pane. Specifically the pane that
// shows the image layers and layer selector.
type LayerController struct {
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
	keybindingPageDown     []keybinding.Key
	keybindingPageUp       []keybinding.Key
}

// NewLayerController creates a new view object attached the the global [gocui] screen object.
func NewLayerController(name string, gui *gocui.Gui, layers []image.Layer) (controller *LayerController) {
	controller = new(LayerController)

	// populate main fields
	controller.Name = name
	controller.gui = gui
	controller.Layers = layers

	switch mode := viper.GetBool("layer.show-aggregated-changes"); mode {
	case true:
		controller.CompareMode = CompareAll
	case false:
		controller.CompareMode = CompareLayer
	default:
		utils.PrintAndExit(fmt.Sprintf("unknown layer.show-aggregated-changes value: %v", mode))
	}

	var err error
	controller.keybindingCompareAll, err = keybinding.ParseAll(viper.GetString("keybinding.compare-all"))
	if err != nil {
		logrus.Error(err)
	}

	controller.keybindingCompareLayer, err = keybinding.ParseAll(viper.GetString("keybinding.compare-layer"))
	if err != nil {
		logrus.Error(err)
	}

	controller.keybindingPageUp, err = keybinding.ParseAll(viper.GetString("keybinding.page-up"))
	if err != nil {
		logrus.Error(err)
	}

	controller.keybindingPageDown, err = keybinding.ParseAll(viper.GetString("keybinding.page-down"))
	if err != nil {
		logrus.Error(err)
	}

	return controller
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (controller *LayerController) Setup(v *gocui.View, header *gocui.View) error {

	// set controller options
	controller.view = v
	controller.view.Editable = false
	controller.view.Wrap = false
	controller.view.Frame = false

	controller.header = header
	controller.header.Editable = false
	controller.header.Wrap = false
	controller.header.Frame = false

	// set keybindings
	if err := controller.gui.SetKeybinding(controller.Name, gocui.KeyArrowDown, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return controller.CursorDown() }); err != nil {
		return err
	}
	if err := controller.gui.SetKeybinding(controller.Name, gocui.KeyArrowUp, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return controller.CursorUp() }); err != nil {
		return err
	}
	if err := controller.gui.SetKeybinding(controller.Name, gocui.KeyArrowRight, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return controller.CursorDown() }); err != nil {
		return err
	}
	if err := controller.gui.SetKeybinding(controller.Name, gocui.KeyArrowLeft, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return controller.CursorUp() }); err != nil {
		return err
	}

	for _, key := range controller.keybindingPageUp {
		if err := controller.gui.SetKeybinding(controller.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return controller.PageUp() }); err != nil {
			return err
		}
	}
	for _, key := range controller.keybindingPageDown {
		if err := controller.gui.SetKeybinding(controller.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return controller.PageDown() }); err != nil {
			return err
		}
	}

	for _, key := range controller.keybindingCompareLayer {
		if err := controller.gui.SetKeybinding(controller.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return controller.setCompareMode(CompareLayer) }); err != nil {
			return err
		}
	}

	for _, key := range controller.keybindingCompareAll {
		if err := controller.gui.SetKeybinding(controller.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return controller.setCompareMode(CompareAll) }); err != nil {
			return err
		}
	}

	return controller.Render()
}

// height obtains the height of the current pane (taking into account the lost space due to the header).
func (controller *LayerController) height() uint {
	_, height := controller.view.Size()
	return uint(height - 1)
}

// IsVisible indicates if the layer view pane is currently initialized.
func (controller *LayerController) IsVisible() bool {
	if controller == nil {
		return false
	}
	return true
}

// PageDown moves to next page putting the cursor on top
func (controller *LayerController) PageDown() error {
	step := int(controller.height()) + 1
	targetLayerIndex := controller.LayerIndex + step

	if targetLayerIndex > len(controller.Layers) {
		step -= targetLayerIndex - (len(controller.Layers) - 1)
		targetLayerIndex = controller.LayerIndex + step
	}

	if step > 0 {
		err := CursorStep(controller.gui, controller.view, step)
		if err == nil {
			controller.SetCursor(controller.LayerIndex + step)
		}
	}
	return nil
}

// PageUp moves to previous page putting the cursor on top
func (controller *LayerController) PageUp() error {
	step := int(controller.height()) + 1
	targetLayerIndex := controller.LayerIndex - step

	if targetLayerIndex < 0 {
		step += targetLayerIndex
		targetLayerIndex = controller.LayerIndex - step
	}

	if step > 0 {
		err := CursorStep(controller.gui, controller.view, -step)
		if err == nil {
			controller.SetCursor(controller.LayerIndex - step)
		}
	}
	return nil
}

// CursorDown moves the cursor down in the layer pane (selecting a higher layer).
func (controller *LayerController) CursorDown() error {
	if controller.LayerIndex < len(controller.Layers) {
		err := CursorDown(controller.gui, controller.view)
		if err == nil {
			controller.SetCursor(controller.LayerIndex + 1)
		}
	}
	return nil
}

// CursorUp moves the cursor up in the layer pane (selecting a lower layer).
func (controller *LayerController) CursorUp() error {
	if controller.LayerIndex > 0 {
		err := CursorUp(controller.gui, controller.view)
		if err == nil {
			controller.SetCursor(controller.LayerIndex - 1)
		}
	}
	return nil
}

// SetCursor resets the cursor and orients the file tree view based on the given layer index.
func (controller *LayerController) SetCursor(layer int) error {
	controller.LayerIndex = layer
	Controllers.Tree.setTreeByLayer(controller.getCompareIndexes())
	Controllers.Details.Render()
	controller.Render()

	return nil
}

// currentLayer returns the Layer object currently selected.
func (controller *LayerController) currentLayer() image.Layer {
	return controller.Layers[(len(controller.Layers)-1)-controller.LayerIndex]
}

// setCompareMode switches the layer comparison between a single-layer comparison to an aggregated comparison.
func (controller *LayerController) setCompareMode(compareMode CompareType) error {
	controller.CompareMode = compareMode
	Update()
	Render()
	return Controllers.Tree.setTreeByLayer(controller.getCompareIndexes())
}

// getCompareIndexes determines the layer boundaries to use for comparison (based on the current compare mode)
func (controller *LayerController) getCompareIndexes() (bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) {
	bottomTreeStart = controller.CompareStartIndex
	topTreeStop = controller.LayerIndex

	if controller.LayerIndex == controller.CompareStartIndex {
		bottomTreeStop = controller.LayerIndex
		topTreeStart = controller.LayerIndex
	} else if controller.CompareMode == CompareLayer {
		bottomTreeStop = controller.LayerIndex - 1
		topTreeStart = controller.LayerIndex
	} else {
		bottomTreeStop = controller.CompareStartIndex
		topTreeStart = controller.CompareStartIndex + 1
	}

	return bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop
}

// renderCompareBar returns the formatted string for the given layer.
func (controller *LayerController) renderCompareBar(layerIdx int) string {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop := controller.getCompareIndexes()
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
func (controller *LayerController) Update() error {
	controller.ImageSize = 0
	for idx := 0; idx < len(controller.Layers); idx++ {
		controller.ImageSize += controller.Layers[idx].Size()
	}
	return nil
}

// Render flushes the state objects to the screen. The layers pane reports:
// 1. the layers of the image + metadata
// 2. the current selected image
func (controller *LayerController) Render() error {

	// indicate when selected
	title := "Layers"
	if controller.gui.CurrentView() == controller.view {
		title = "● " + title
	}

	controller.gui.Update(func(g *gocui.Gui) error {
		// update header
		controller.header.Clear()
		width, _ := g.Size()
		headerStr := fmt.Sprintf("[%s]%s\n", title, strings.Repeat("─", width*2))
		// headerStr += fmt.Sprintf("Cmp "+image.LayerFormat, "Layer Digest", "Size", "Command")
		headerStr += fmt.Sprintf("Cmp"+image.LayerFormat, "Size", "Command")
		fmt.Fprintln(controller.header, Formatting.Header(vtclean.Clean(headerStr, false)))

		// update contents
		controller.view.Clear()
		for revIdx := len(controller.Layers) - 1; revIdx >= 0; revIdx-- {
			layer := controller.Layers[revIdx]
			idx := (len(controller.Layers) - 1) - revIdx

			layerStr := layer.String()
			compareBar := controller.renderCompareBar(idx)

			if idx == controller.LayerIndex {
				fmt.Fprintln(controller.view, compareBar+" "+Formatting.Selected(layerStr))
			} else {
				fmt.Fprintln(controller.view, compareBar+" "+layerStr)
			}

		}
		return nil
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected.
func (controller *LayerController) KeyHelp() string {
	return renderStatusOption(controller.keybindingCompareLayer[0].String(), "Show layer changes", controller.CompareMode == CompareLayer) +
		renderStatusOption(controller.keybindingCompareAll[0].String(), "Show aggregated changes", controller.CompareMode == CompareAll)
}
