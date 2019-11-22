package view

import (
	"fmt"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/key"
	"github.com/wagoodman/dive/runtime/ui/viewmodel"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/lunixbochs/vtclean"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type LayerChangeListener func(viewmodel.LayerSelection) error

// Layer holds the UI objects and data models for populating the lower-left pane. Specifically the pane that
// shows the image layers and layer selector.
type Layer struct {
	name              string
	gui               *gocui.Gui
	view              *gocui.View
	header            *gocui.View
	LayerIndex        int
	Layers            []*image.Layer
	CompareMode       CompareType
	CompareStartIndex int

	listeners []LayerChangeListener

	helpKeys []*key.Binding
}

// newLayerView creates a new view object attached the the global [gocui] screen object.
func newLayerView(name string, gui *gocui.Gui, layers []*image.Layer) (controller *Layer, err error) {
	controller = new(Layer)

	controller.listeners = make([]LayerChangeListener, 0)

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

func (v *Layer) AddLayerChangeListener(listener ...LayerChangeListener) {
	v.listeners = append(v.listeners, listener...)
}

func (v *Layer) notifyLayerChangeListeners() error {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop := v.getCompareIndexes()
	selection := viewmodel.LayerSelection{
		Layer:           v.CurrentLayer(),
		BottomTreeStart: bottomTreeStart,
		BottomTreeStop:  bottomTreeStop,
		TopTreeStart:    topTreeStart,
		TopTreeStop:     topTreeStop,
	}
	for _, listener := range v.listeners {
		err := listener(selection)
		if err != nil {
			logrus.Errorf("notifyLayerChangeListeners error: %+v", err)
			return err
		}
	}
	return nil
}

func (v *Layer) Name() string {
	return v.name
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (v *Layer) Setup(view *gocui.View, header *gocui.View) error {
	logrus.Debugf("view.Setup() %s", v.Name())

	// set controller options
	v.view = view
	v.view.Editable = false
	v.view.Wrap = false
	v.view.Frame = false

	v.header = header
	v.header.Editable = false
	v.header.Wrap = false
	v.header.Frame = false

	var infos = []key.BindingInfo{
		{
			ConfigKeys: []string{"keybinding.compare-layer"},
			OnAction:   func() error { return v.setCompareMode(CompareLayer) },
			IsSelected: func() bool { return v.CompareMode == CompareLayer },
			Display:    "Show layer changes",
		},
		{
			ConfigKeys: []string{"keybinding.compare-all"},
			OnAction:   func() error { return v.setCompareMode(CompareAll) },
			IsSelected: func() bool { return v.CompareMode == CompareAll },
			Display:    "Show aggregated changes",
		},
		{
			Key:      gocui.KeyArrowDown,
			Modifier: gocui.ModNone,
			OnAction: v.CursorDown,
		},
		{
			Key:      gocui.KeyArrowUp,
			Modifier: gocui.ModNone,
			OnAction: v.CursorUp,
		},
		{
			Key:      gocui.KeyArrowLeft,
			Modifier: gocui.ModNone,
			OnAction: v.CursorUp,
		},
		{
			Key:      gocui.KeyArrowRight,
			Modifier: gocui.ModNone,
			OnAction: v.CursorDown,
		},
		{
			ConfigKeys: []string{"keybinding.page-up"},
			OnAction:   v.PageUp,
		},
		{
			ConfigKeys: []string{"keybinding.page-down"},
			OnAction:   v.PageDown,
		},
	}

	helpKeys, err := key.GenerateBindings(v.gui, v.name, infos)
	if err != nil {
		return err
	}
	v.helpKeys = helpKeys

	return v.Render()
}

// height obtains the height of the current pane (taking into account the lost space due to the header).
func (v *Layer) height() uint {
	_, height := v.view.Size()
	return uint(height - 1)
}

// IsVisible indicates if the layer view pane is currently initialized.
func (v *Layer) IsVisible() bool {
	return v != nil
}

// PageDown moves to next page putting the cursor on top
func (v *Layer) PageDown() error {
	step := int(v.height()) + 1
	targetLayerIndex := v.LayerIndex + step

	if targetLayerIndex > len(v.Layers) {
		step -= targetLayerIndex - (len(v.Layers) - 1)
	}

	if step > 0 {
		err := CursorStep(v.gui, v.view, step)
		if err == nil {
			return v.SetCursor(v.LayerIndex + step)
		}
	}
	return nil
}

// PageUp moves to previous page putting the cursor on top
func (v *Layer) PageUp() error {
	step := int(v.height()) + 1
	targetLayerIndex := v.LayerIndex - step

	if targetLayerIndex < 0 {
		step += targetLayerIndex
	}

	if step > 0 {
		err := CursorStep(v.gui, v.view, -step)
		if err == nil {
			return v.SetCursor(v.LayerIndex - step)
		}
	}
	return nil
}

// CursorDown moves the cursor down in the layer pane (selecting a higher layer).
func (v *Layer) CursorDown() error {
	if v.LayerIndex < len(v.Layers) {
		err := CursorDown(v.gui, v.view)
		if err == nil {
			return v.SetCursor(v.LayerIndex + 1)
		}
	}
	return nil
}

// CursorUp moves the cursor up in the layer pane (selecting a lower layer).
func (v *Layer) CursorUp() error {
	if v.LayerIndex > 0 {
		err := CursorUp(v.gui, v.view)
		if err == nil {
			return v.SetCursor(v.LayerIndex - 1)
		}
	}
	return nil
}

// SetCursor resets the cursor and orients the file tree view based on the given layer index.
func (v *Layer) SetCursor(layer int) error {
	v.LayerIndex = layer
	err := v.notifyLayerChangeListeners()
	if err != nil {
		return err
	}

	return v.Render()
}

// CurrentLayer returns the Layer object currently selected.
func (v *Layer) CurrentLayer() *image.Layer {
	return v.Layers[v.LayerIndex]
}

// setCompareMode switches the layer comparison between a single-layer comparison to an aggregated comparison.
func (v *Layer) setCompareMode(compareMode CompareType) error {
	v.CompareMode = compareMode
	return v.notifyLayerChangeListeners()
}

// getCompareIndexes determines the layer boundaries to use for comparison (based on the current compare mode)
func (v *Layer) getCompareIndexes() (bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) {
	bottomTreeStart = v.CompareStartIndex
	topTreeStop = v.LayerIndex

	if v.LayerIndex == v.CompareStartIndex {
		bottomTreeStop = v.LayerIndex
		topTreeStart = v.LayerIndex
	} else if v.CompareMode == CompareLayer {
		bottomTreeStop = v.LayerIndex - 1
		topTreeStart = v.LayerIndex
	} else {
		bottomTreeStop = v.CompareStartIndex
		topTreeStart = v.CompareStartIndex + 1
	}

	return bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop
}

// renderCompareBar returns the formatted string for the given layer.
func (v *Layer) renderCompareBar(layerIdx int) string {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop := v.getCompareIndexes()
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
func (v *Layer) Update() error {
	return nil
}

// Render flushes the state objects to the screen. The layers pane reports:
// 1. the layers of the image + metadata
// 2. the current selected image
func (v *Layer) Render() error {
	logrus.Debugf("view.Render() %s", v.Name())

	// indicate when selected
	title := "Layers"
	if v.gui.CurrentView() == v.view {
		title = "● " + title
	}

	v.gui.Update(func(g *gocui.Gui) error {
		// update header
		v.header.Clear()
		width, _ := g.Size()
		headerStr := fmt.Sprintf("[%s]%s\n", title, strings.Repeat("─", width*2))
		headerStr += fmt.Sprintf("Cmp"+image.LayerFormat, "Size", "Command")
		_, err := fmt.Fprintln(v.header, format.Header(vtclean.Clean(headerStr, false)))
		if err != nil {
			return err
		}

		// update contents
		v.view.Clear()
		for idx, layer := range v.Layers {

			layerStr := layer.String()
			compareBar := v.renderCompareBar(idx)

			if idx == v.LayerIndex {
				_, err = fmt.Fprintln(v.view, compareBar+" "+format.Selected(layerStr))
			} else {
				_, err = fmt.Fprintln(v.view, compareBar+" "+layerStr)
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
func (v *Layer) KeyHelp() string {
	var help string
	for _, binding := range v.helpKeys {
		help += binding.RenderKeyHelp()
	}
	return help
}
