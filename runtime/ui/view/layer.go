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

// NewLayerView creates a new view object attached the the global [gocui] screen object.
func NewLayerView(name string, gui *gocui.Gui, layers []*image.Layer) (controller *Layer, err error) {
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

func (c *Layer) AddLayerChangeListener(listener ...LayerChangeListener) {
	c.listeners = append(c.listeners, listener...)
}

func (c *Layer) notifyLayerChangeListeners() error {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop := c.getCompareIndexes()
	selection := viewmodel.LayerSelection{
		Layer:           c.CurrentLayer(),
		BottomTreeStart: bottomTreeStart,
		BottomTreeStop:  bottomTreeStop,
		TopTreeStart:    topTreeStart,
		TopTreeStop:     topTreeStop,
	}
	for _, listener := range c.listeners {
		err := listener(selection)
		if err != nil {
			logrus.Errorf("notifyLayerChangeListeners error: %+v", err)
			return err
		}
	}
	return nil
}

func (c *Layer) Height() int {
	return HeightFull
}

func (c *Layer) Width() int {
	return WidthFull
}


func (c *Layer) Name() string {
	return c.name
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (c *Layer) Setup(v *gocui.View, header *gocui.View) error {

	// set controller options
	c.view = v
	c.view.Editable = false
	c.view.Wrap = false
	c.view.Frame = false

	c.header = header
	c.header.Editable = false
	c.header.Wrap = false
	c.header.Frame = false

	var infos = []key.BindingInfo{
		{
			ConfigKeys: []string{"keybinding.compare-layer"},
			OnAction:   func() error { return c.setCompareMode(CompareLayer) },
			IsSelected: func() bool { return c.CompareMode == CompareLayer },
			Display:    "Show layer changes",
		},
		{
			ConfigKeys: []string{"keybinding.compare-all"},
			OnAction:   func() error { return c.setCompareMode(CompareAll) },
			IsSelected: func() bool { return c.CompareMode == CompareAll },
			Display:    "Show aggregated changes",
		},
		{
			Key:      gocui.KeyArrowDown,
			Modifier: gocui.ModNone,
			OnAction: c.CursorDown,
		},
		{
			Key:      gocui.KeyArrowUp,
			Modifier: gocui.ModNone,
			OnAction: c.CursorUp,
		},
		{
			Key:      gocui.KeyArrowLeft,
			Modifier: gocui.ModNone,
			OnAction: c.CursorUp,
		},
		{
			Key:      gocui.KeyArrowRight,
			Modifier: gocui.ModNone,
			OnAction: c.CursorDown,
		},
		{
			ConfigKeys: []string{"keybinding.page-up"},
			OnAction:   c.PageUp,
		},
		{
			ConfigKeys: []string{"keybinding.page-down"},
			OnAction:   c.PageDown,
		},
	}

	helpKeys, err := key.GenerateBindings(c.gui, c.name, infos)
	if err != nil {
		return err
	}
	c.helpKeys = helpKeys

	return c.Render()
}

// height obtains the height of the current pane (taking into account the lost space due to the header).
func (c *Layer) height() uint {
	_, height := c.view.Size()
	return uint(height - 1)
}

// IsVisible indicates if the layer view pane is currently initialized.
func (c *Layer) IsVisible() bool {
	return c != nil
}

// PageDown moves to next page putting the cursor on top
func (c *Layer) PageDown() error {
	step := int(c.height()) + 1
	targetLayerIndex := c.LayerIndex + step

	if targetLayerIndex > len(c.Layers) {
		step -= targetLayerIndex - (len(c.Layers) - 1)
	}

	if step > 0 {
		err := CursorStep(c.gui, c.view, step)
		if err == nil {
			return c.SetCursor(c.LayerIndex + step)
		}
	}
	return nil
}

// PageUp moves to previous page putting the cursor on top
func (c *Layer) PageUp() error {
	step := int(c.height()) + 1
	targetLayerIndex := c.LayerIndex - step

	if targetLayerIndex < 0 {
		step += targetLayerIndex
	}

	if step > 0 {
		err := CursorStep(c.gui, c.view, -step)
		if err == nil {
			return c.SetCursor(c.LayerIndex - step)
		}
	}
	return nil
}

// CursorDown moves the cursor down in the layer pane (selecting a higher layer).
func (c *Layer) CursorDown() error {
	if c.LayerIndex < len(c.Layers) {
		err := CursorDown(c.gui, c.view)
		if err == nil {
			return c.SetCursor(c.LayerIndex + 1)
		}
	}
	return nil
}

// CursorUp moves the cursor up in the layer pane (selecting a lower layer).
func (c *Layer) CursorUp() error {
	if c.LayerIndex > 0 {
		err := CursorUp(c.gui, c.view)
		if err == nil {
			return c.SetCursor(c.LayerIndex - 1)
		}
	}
	return nil
}

// SetCursor resets the cursor and orients the file tree view based on the given layer index.
func (c *Layer) SetCursor(layer int) error {
	c.LayerIndex = layer
	err := c.notifyLayerChangeListeners()
	if err != nil {
		return err
	}

	return c.Render()
}

// CurrentLayer returns the Layer object currently selected.
func (c *Layer) CurrentLayer() *image.Layer {
	return c.Layers[c.LayerIndex]
}

// setCompareMode switches the layer comparison between a single-layer comparison to an aggregated comparison.
func (c *Layer) setCompareMode(compareMode CompareType) error {
	c.CompareMode = compareMode
	return c.notifyLayerChangeListeners()
}

// getCompareIndexes determines the layer boundaries to use for comparison (based on the current compare mode)
func (c *Layer) getCompareIndexes() (bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) {
	bottomTreeStart = c.CompareStartIndex
	topTreeStop = c.LayerIndex

	if c.LayerIndex == c.CompareStartIndex {
		bottomTreeStop = c.LayerIndex
		topTreeStart = c.LayerIndex
	} else if c.CompareMode == CompareLayer {
		bottomTreeStop = c.LayerIndex - 1
		topTreeStart = c.LayerIndex
	} else {
		bottomTreeStop = c.CompareStartIndex
		topTreeStart = c.CompareStartIndex + 1
	}

	return bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop
}

// renderCompareBar returns the formatted string for the given layer.
func (c *Layer) renderCompareBar(layerIdx int) string {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop := c.getCompareIndexes()
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
func (c *Layer) Update() error {
	return nil
}

// Render flushes the state objects to the screen. The layers pane reports:
// 1. the layers of the image + metadata
// 2. the current selected image
func (c *Layer) Render() error {

	// indicate when selected
	title := "Layers"
	if c.gui.CurrentView() == c.view {
		title = "● " + title
	}

	c.gui.Update(func(g *gocui.Gui) error {
		// update header
		c.header.Clear()
		width, _ := g.Size()
		headerStr := fmt.Sprintf("[%s]%s\n", title, strings.Repeat("─", width*2))
		headerStr += fmt.Sprintf("Cmp"+image.LayerFormat, "Size", "Command")
		_, err := fmt.Fprintln(c.header, format.Header(vtclean.Clean(headerStr, false)))
		if err != nil {
			return err
		}

		// update contents
		c.view.Clear()
		for idx, layer := range c.Layers {

			layerStr := layer.String()
			compareBar := c.renderCompareBar(idx)

			if idx == c.LayerIndex {
				_, err = fmt.Fprintln(c.view, compareBar+" "+format.Selected(layerStr))
			} else {
				_, err = fmt.Fprintln(c.view, compareBar+" "+layerStr)
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
func (c *Layer) KeyHelp() string {
	var help string
	for _, binding := range c.helpKeys {
		help += binding.RenderKeyHelp()
	}
	return help
}
