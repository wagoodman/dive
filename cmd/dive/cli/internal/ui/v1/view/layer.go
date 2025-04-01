package view

import (
	"fmt"
	"github.com/anchore/go-logger"
	"github.com/awesome-gocui/gocui"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/format"
	key2 "github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/key"
	viewmodel2 "github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/viewmodel"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/internal/log"
)

// Layer holds the UI objects and data models for populating the lower-left pane.
// Specifically the pane that shows the image layers and layer selector.
type Layer struct {
	name                  string
	gui                   *gocui.Gui
	body                  *gocui.View
	header                *gocui.View
	vm                    *viewmodel2.LayerSetState
	kb                    key2.Bindings
	logger                logger.Logger
	constrainedRealEstate bool

	listeners []LayerChangeListener

	helpKeys []*key2.Binding
}

// newLayerView creates a new view object attached the global [gocui] screen object.
func newLayerView(gui *gocui.Gui, cfg v1.Config) (c *Layer, err error) {
	c = new(Layer)

	c.logger = log.Nested("ui", "layer")
	c.listeners = make([]LayerChangeListener, 0)

	// populate main fields
	c.name = "layer"
	c.gui = gui
	c.kb = cfg.Preferences.KeyBindings

	var compareMode viewmodel2.LayerCompareMode

	switch mode := cfg.Preferences.ShowAggregatedLayerChanges; mode {
	case true:
		compareMode = viewmodel2.CompareAllLayers
	case false:
		compareMode = viewmodel2.CompareSingleLayer
	default:
		return nil, fmt.Errorf("unknown layer.show-aggregated-changes value: %v", mode)
	}

	c.vm = viewmodel2.NewLayerSetState(cfg.Analysis.Layers, compareMode)

	return c, err
}

func (v *Layer) AddLayerChangeListener(listener ...LayerChangeListener) {
	v.listeners = append(v.listeners, listener...)
}

func (v *Layer) notifyLayerChangeListeners() error {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop := v.vm.GetCompareIndexes()
	selection := viewmodel2.LayerSelection{
		Layer:           v.CurrentLayer(),
		BottomTreeStart: bottomTreeStart,
		BottomTreeStop:  bottomTreeStop,
		TopTreeStart:    topTreeStart,
		TopTreeStop:     topTreeStop,
	}
	for _, listener := range v.listeners {
		err := listener(selection)
		if err != nil {
			return fmt.Errorf("error notifying layer change listeners: %w", err)
		}
	}
	// this is hacky, and I do not like it
	if layerDetails, err := v.gui.View("layerDetails"); err == nil {
		if err := layerDetails.SetCursor(0, 0); err != nil {
			v.logger.Debug("Couldn't set cursor to 0,0 for layerDetails")
		}
	}
	return nil
}

func (v *Layer) Name() string {
	return v.name
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (v *Layer) Setup(body *gocui.View, header *gocui.View) error {
	v.logger.Trace("Setup()")

	// set controller options
	v.body = body
	v.body.Editable = false
	v.body.Wrap = false
	v.body.Frame = false

	v.header = header
	v.header.Editable = false
	v.header.Wrap = false
	v.header.Frame = false

	var infos = []key2.BindingInfo{
		{
			Config:     v.kb.Layer.CompareLayer,
			OnAction:   func() error { return v.setCompareMode(viewmodel2.CompareSingleLayer) },
			IsSelected: func() bool { return v.vm.CompareMode == viewmodel2.CompareSingleLayer },
			Display:    "Show layer changes",
		},
		{
			Config:     v.kb.Layer.CompareAll,
			OnAction:   func() error { return v.setCompareMode(viewmodel2.CompareAllLayers) },
			IsSelected: func() bool { return v.vm.CompareMode == viewmodel2.CompareAllLayers },
			Display:    "Show aggregated changes",
		},
		{
			Config:   v.kb.Navigation.Down,
			Modifier: gocui.ModNone,
			OnAction: v.CursorDown,
		},
		{
			Config:   v.kb.Navigation.Up,
			Modifier: gocui.ModNone,
			OnAction: v.CursorUp,
		},
		{
			Config:   v.kb.Navigation.PageUp,
			OnAction: v.PageUp,
		},
		{
			Config:   v.kb.Navigation.PageDown,
			OnAction: v.PageDown,
		},
	}

	helpKeys, err := key2.GenerateBindings(v.gui, v.name, infos)
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

func (v *Layer) CompareMode() viewmodel2.LayerCompareMode {
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

// SetOrigin updates the origin of the layer view pane.
func (v *Layer) SetOrigin(x, y int) error {
	if err := v.body.SetOrigin(x, y); err != nil {
		return err
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
func (v *Layer) setCompareMode(compareMode viewmodel2.LayerCompareMode) error {
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
		v.logger.Debug("constraining layout")
		v.constrainedRealEstate = true
	}
}

func (v *Layer) ExpandLayout() {
	if v.constrainedRealEstate {
		v.logger.Debug("expanding layout")
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
	v.logger.Trace("render()")

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
				return err
			}
		}

		// Adjust origin, if necessary
		maxBodyDisplayHeight := int(v.height())
		if v.vm.LayerIndex > maxBodyDisplayHeight {
			if err := v.SetOrigin(0, v.vm.LayerIndex-maxBodyDisplayHeight); err != nil {
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
