package view

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/key"
	"github.com/wagoodman/dive/runtime/ui/viewmodel"
)

// Layer holds the UI objects and data models for populating the lower-left pane.
// Specifically the pane that shows the image layers and layer selector.
type Layer struct {
	name                  string
	gui                   *gocui.Gui
	body                  *gocui.View
	header                *gocui.View
	vm                    *viewmodel.LayerSetState
	constrainedRealEstate bool

	listeners []LayerChangeListener

	helpKeys []*key.Binding
}

// newLayerView creates a new view object attached the the global [gocui] screen object.
func newLayerView(gui *gocui.Gui, layers []*image.Layer) (controller *Layer, err error) {
	controller = new(Layer)

	controller.listeners = make([]LayerChangeListener, 0)

	// populate main fields
	controller.name = "layer"
	controller.gui = gui

	var compareMode viewmodel.LayerCompareMode

	switch mode := viper.GetBool("layer.show-aggregated-changes"); mode {
	case true:
		compareMode = viewmodel.CompareAllLayers
	case false:
		compareMode = viewmodel.CompareSingleLayer
	default:
		return nil, fmt.Errorf("unknown layer.show-aggregated-changes value: %v", mode)
	}

	controller.vm = viewmodel.NewLayerSetState(layers, compareMode)

	return controller, err
}

func (v *Layer) AddLayerChangeListener(listener ...LayerChangeListener) {
	v.listeners = append(v.listeners, listener...)
}

func (v *Layer) notifyLayerChangeListeners() error {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop := v.vm.GetCompareIndexes()
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
	// this is hacky, and I do not like it
	if layerDetails, err := v.gui.View("layerDetails"); err == nil {
		if err := layerDetails.SetCursor(0, 0); err != nil {
			logrus.Debug("Couldn't set cursor to 0,0 for layerDetails")
		}
	}
	return nil
}

func (v *Layer) Name() string {
	return v.name
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (v *Layer) Setup(body *gocui.View, header *gocui.View) error {
	logrus.Tracef("view.Setup() %s", v.Name())

	// set controller options
	v.body = body
	v.body.Editable = false
	v.body.Wrap = false
	v.body.Frame = false

	v.header = header
	v.header.Editable = false
	v.header.Wrap = false
	v.header.Frame = false

	var infos = []key.BindingInfo{
		{
			ConfigKeys: []string{"keybinding.compare-layer"},
			OnAction:   func() error { return v.setCompareMode(viewmodel.CompareSingleLayer) },
			IsSelected: func() bool { return v.vm.CompareMode == viewmodel.CompareSingleLayer },
			Display:    "Show layer changes",
		},
		{
			ConfigKeys: []string{"keybinding.compare-all"},
			OnAction:   func() error { return v.setCompareMode(viewmodel.CompareAllLayers) },
			IsSelected: func() bool { return v.vm.CompareMode == viewmodel.CompareAllLayers },
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
	_, height := v.body.Size()
	return uint(height - 1)
}

func (v *Layer) CompareMode() viewmodel.LayerCompareMode {
	return v.vm.CompareMode
}

// IsVisible indicates if the layer view pane is currently initialized.
func (v *Layer) IsVisible() bool {
	return v != nil
}

// PageDown moves to next page putting the cursor on top
func (v *Layer) PageDown() error {
	step := int(v.height()) + 1
	targetLayerIndex := v.vm.LayerIndex + step

	if targetLayerIndex > len(v.vm.Layers) {
		step -= targetLayerIndex - (len(v.vm.Layers) - 1)
	}

	if step > 0 {
		// err := CursorStep(v.gui, v.body, step)
		err := error(nil)
		if err == nil {
			return v.SetCursor(v.vm.LayerIndex + step)
		}
	}
	return nil
}

// PageUp moves to previous page putting the cursor on top
func (v *Layer) PageUp() error {
	step := int(v.height()) + 1
	targetLayerIndex := v.vm.LayerIndex - step

	if targetLayerIndex < 0 {
		step += targetLayerIndex
	}

	if step > 0 {
		// err := CursorStep(v.gui, v.body, -step)
		err := error(nil)
		if err == nil {
			return v.SetCursor(v.vm.LayerIndex - step)
		}
	}
	return nil
}

// CursorDown moves the cursor down in the layer pane (selecting a higher layer).
func (v *Layer) CursorDown() error {
	if v.vm.LayerIndex < len(v.vm.Layers)-1 {
		// err := CursorDown(v.gui, v.body)
		err := error(nil)
		if err == nil {
			return v.SetCursor(v.vm.LayerIndex + 1)
		}
	}
	return nil
}

// CursorUp moves the cursor up in the layer pane (selecting a lower layer).
func (v *Layer) CursorUp() error {
	if v.vm.LayerIndex > 0 {
		// err := CursorUp(v.gui, v.body)
		err := error(nil)
		if err == nil {
			return v.SetCursor(v.vm.LayerIndex - 1)
		}
	}
	return nil
}

// SetCursor resets the cursor and orients the file tree view based on the given layer index.
func (v *Layer) SetCursor(layer int) error {
	v.vm.LayerIndex = layer
	err := v.notifyLayerChangeListeners()
	if err != nil {
		return err
	}

	return v.Render()
}

// CurrentLayer returns the Layer object currently selected.
func (v *Layer) CurrentLayer() *image.Layer {
	return v.vm.Layers[v.vm.LayerIndex]
}

// setCompareMode switches the layer comparison between a single-layer comparison to an aggregated comparison.
func (v *Layer) setCompareMode(compareMode viewmodel.LayerCompareMode) error {
	v.vm.CompareMode = compareMode
	return v.notifyLayerChangeListeners()
}

// renderCompareBar returns the formatted string for the given layer.
func (v *Layer) renderCompareBar(layerIdx int) string {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop := v.vm.GetCompareIndexes()
	result := "  "

	if layerIdx >= bottomTreeStart && layerIdx <= bottomTreeStop {
		result = format.CompareBottom("  ")
	}
	if layerIdx >= topTreeStart && layerIdx <= topTreeStop {
		result = format.CompareTop("  ")
	}

	return result
}

func (v *Layer) ConstrainLayout() {
	if !v.constrainedRealEstate {
		logrus.Debugf("constraining layer layout")
		v.constrainedRealEstate = true
	}
}

func (v *Layer) ExpandLayout() {
	if v.constrainedRealEstate {
		logrus.Debugf("expanding layer layout")
		v.constrainedRealEstate = false
	}
}

// OnLayoutChange is called whenever the screen dimensions are changed
func (v *Layer) OnLayoutChange() error {
	err := v.Update()
	if err != nil {
		return err
	}
	return v.Render()
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (v *Layer) Update() error {
	return nil
}

// Render flushes the state objects to the screen. The layers pane reports:
// 1. the layers of the image + metadata
// 2. the current selected image
func (v *Layer) Render() error {
	logrus.Tracef("view.Render() %s", v.Name())

	// indicate when selected
	title := "Layers"
	isSelected := v.gui.CurrentView() == v.body

	v.gui.Update(func(g *gocui.Gui) error {
		var err error
		// update header
		v.header.Clear()
		width, _ := g.Size()
		if v.constrainedRealEstate {
			headerStr := format.RenderNoHeader(width, isSelected)
			headerStr += "\nLayer"
			_, err := fmt.Fprintln(v.header, headerStr)
			if err != nil {
				return err
			}
		} else {
			headerStr := format.RenderHeader(title, width, isSelected)
			headerStr += fmt.Sprintf("Cmp"+image.LayerFormat, "Size", "Command")
			_, err := fmt.Fprintln(v.header, headerStr)
			if err != nil {
				return err
			}
		}

		// update contents
		v.body.Clear()
		for idx, layer := range v.vm.Layers {
			var layerStr string
			if v.constrainedRealEstate {
				layerStr = fmt.Sprintf("%-4d", layer.Index)
			} else {
				layerStr = layer.String()
			}

			compareBar := v.renderCompareBar(idx)

			if idx == v.vm.LayerIndex {
				_, err = fmt.Fprintln(v.body, compareBar+" "+format.Selected(layerStr))
			} else {
				_, err = fmt.Fprintln(v.body, compareBar+" "+layerStr)
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

func (v *Layer) LayerCount() int {
	return len(v.vm.Layers)
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected.
func (v *Layer) KeyHelp() string {
	var help string
	for _, binding := range v.helpKeys {
		help += binding.RenderKeyHelp()
	}
	return help
}
