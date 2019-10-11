package ui

import (
	"fmt"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/key"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/lunixbochs/vtclean"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// layerController holds the UI objects and data models for populating the lower-left pane. Specifically the pane that
// shows the image layers and layer selector.
type layerController struct {
	name              string
	gui               *gocui.Gui
	view              *gocui.View
	header            *gocui.View
	LayerIndex        int
	Layers            []*image.Layer
	CompareMode       CompareType
	CompareStartIndex int
	ImageSize         uint64

	helpKeys []*key.Binding
}

// newLayerController creates a new view object attached the the global [gocui] screen object.
func newLayerController(name string, gui *gocui.Gui, layers []*image.Layer) (controller *layerController, err error) {
	controller = new(layerController)

	// populate main fields
	controller.name = name
	controller.gui = gui
	controller.Layers = layers

	switch mode := viper.GetBool("layer.show-aggregated-changes"); mode {
	case true:
		controller.CompareMode = CompareAll
	case false:
		controller.CompareMode = CompareLayer
	default:
		return nil, fmt.Errorf("unknown layer.show-aggregated-changes value: %v", mode)
	}

	return controller, err
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (controller *layerController) Setup(v *gocui.View, header *gocui.View) error {

	// set controller options
	controller.view = v
	controller.view.Editable = false
	controller.view.Wrap = false
	controller.view.Frame = false

	controller.header = header
	controller.header.Editable = false
	controller.header.Wrap = false
	controller.header.Frame = false


	var infos = []key.BindingInfo{
		{
			ConfigKeys: []string{"keybinding.compare-layer"},
			OnAction:   func() error { return  controller.setCompareMode(CompareLayer) },
			IsSelected: func() bool { return controller.CompareMode == CompareLayer },
			Display:    "Show layer changes",
		},
		{
			ConfigKeys: []string{"keybinding.compare-all"},
			OnAction:   func() error { return  controller.setCompareMode(CompareAll) },
			IsSelected: func() bool { return controller.CompareMode == CompareAll },
			Display:    "Show aggregated changes",
		},
		{
			Key:      gocui.KeyArrowDown,
			Modifier: gocui.ModNone,
			OnAction: controller.CursorDown,
		},
		{
			Key:      gocui.KeyArrowUp,
			Modifier: gocui.ModNone,
			OnAction: controller.CursorUp,
		},
		{
			Key:      gocui.KeyArrowLeft,
			Modifier: gocui.ModNone,
			OnAction: controller.CursorUp,
		},
		{
			Key:      gocui.KeyArrowRight,
			Modifier: gocui.ModNone,
			OnAction: controller.CursorDown,
		},
		{
			ConfigKeys: []string{"keybinding.page-up"},
			OnAction:   controller.PageUp,
		},
		{
			ConfigKeys: []string{"keybinding.page-down"},
			OnAction:   controller.PageDown,
		},
	}

	helpKeys, err := key.GenerateBindings(controller.gui, controller.name, infos)
	if err != nil {
		return err
	}
	controller.helpKeys = helpKeys


	return controller.Render()
}

// height obtains the height of the current pane (taking into account the lost space due to the header).
func (controller *layerController) height() uint {
	_, height := controller.view.Size()
	return uint(height - 1)
}

// IsVisible indicates if the layer view pane is currently initialized.
func (controller *layerController) IsVisible() bool {
	return controller != nil
}

// PageDown moves to next page putting the cursor on top
func (controller *layerController) PageDown() error {
	step := int(controller.height()) + 1
	targetLayerIndex := controller.LayerIndex + step

	if targetLayerIndex > len(controller.Layers) {
		step -= targetLayerIndex - (len(controller.Layers) - 1)
	}

	if step > 0 {
		err := CursorStep(controller.gui, controller.view, step)
		if err == nil {
			return controller.SetCursor(controller.LayerIndex + step)
		}
	}
	return nil
}

// PageUp moves to previous page putting the cursor on top
func (controller *layerController) PageUp() error {
	step := int(controller.height()) + 1
	targetLayerIndex := controller.LayerIndex - step

	if targetLayerIndex < 0 {
		step += targetLayerIndex
	}

	if step > 0 {
		err := CursorStep(controller.gui, controller.view, -step)
		if err == nil {
			return controller.SetCursor(controller.LayerIndex - step)
		}
	}
	return nil
}

// CursorDown moves the cursor down in the layer pane (selecting a higher layer).
func (controller *layerController) CursorDown() error {
	if controller.LayerIndex < len(controller.Layers) {
		err := CursorDown(controller.gui, controller.view)
		if err == nil {
			return controller.SetCursor(controller.LayerIndex + 1)
		}
	}
	return nil
}

// CursorUp moves the cursor up in the layer pane (selecting a lower layer).
func (controller *layerController) CursorUp() error {
	if controller.LayerIndex > 0 {
		err := CursorUp(controller.gui, controller.view)
		if err == nil {
			return controller.SetCursor(controller.LayerIndex - 1)
		}
	}
	return nil
}

// SetCursor resets the cursor and orients the file tree view based on the given layer index.
func (controller *layerController) SetCursor(layer int) error {
	controller.LayerIndex = layer
	err := controllers.Tree.setTreeByLayer(controller.getCompareIndexes())
	if err != nil {
		return err
	}

	_ = controllers.Details.Render()

	return controller.Render()
}

// currentLayer returns the Layer object currently selected.
func (controller *layerController) currentLayer() *image.Layer {
	return controller.Layers[controller.LayerIndex]
}

// setCompareMode switches the layer comparison between a single-layer comparison to an aggregated comparison.
func (controller *layerController) setCompareMode(compareMode CompareType) error {
	controller.CompareMode = compareMode
	err := UpdateAndRender()
	if err != nil {
		logrus.Errorf("unable to set compare mode: %+v", err)
		return err
	}
	return controllers.Tree.setTreeByLayer(controller.getCompareIndexes())
}

// getCompareIndexes determines the layer boundaries to use for comparison (based on the current compare mode)
func (controller *layerController) getCompareIndexes() (bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) {
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
func (controller *layerController) renderCompareBar(layerIdx int) string {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop := controller.getCompareIndexes()
	result := "  "

	if layerIdx >= bottomTreeStart && layerIdx <= bottomTreeStop {
		result = format.CompareBottom("  ")
	}
	if layerIdx >= topTreeStart && layerIdx <= topTreeStop {
		result = format.CompareTop("  ")
	}

	return result
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (controller *layerController) Update() error {
	controller.ImageSize = 0
	for idx := 0; idx < len(controller.Layers); idx++ {
		controller.ImageSize += controller.Layers[idx].Size
	}
	return nil
}

// Render flushes the state objects to the screen. The layers pane reports:
// 1. the layers of the image + metadata
// 2. the current selected image
func (controller *layerController) Render() error {

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
		headerStr += fmt.Sprintf("Cmp"+image.LayerFormat, "Size", "Command")
		_, err := fmt.Fprintln(controller.header, format.Header(vtclean.Clean(headerStr, false)))
		if err != nil {
			return err
		}

		// update contents
		controller.view.Clear()
		for idx, layer := range controller.Layers {

			layerStr := layer.String()
			compareBar := controller.renderCompareBar(idx)

			if idx == controller.LayerIndex {
				_, err = fmt.Fprintln(controller.view, compareBar+" "+format.Selected(layerStr))
			} else {
				_, err = fmt.Fprintln(controller.view, compareBar+" "+layerStr)
			}

			if err != nil {
				logrus.Debug("unable to write to buffer: ", err)
				return err
			}

		}
		return nil
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected.
func (controller *layerController) KeyHelp() string {
	var help string
	for _, binding := range controller.helpKeys {
		help += binding.RenderKeyHelp()
	}
	return help
}
