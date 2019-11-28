package view

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/key"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/jroimartin/gocui"
)

// Details holds the UI objects and data models for populating the lower-left pane. Specifically the pane that
// shows the layer details and image statistics.
type Details struct {
	name           string
	gui            *gocui.Gui
	view           *gocui.View
	header         *gocui.View
	efficiency     float64
	inefficiencies filetree.EfficiencySlice
	imageSize      uint64

	currentLayer *image.Layer
}

// newDetailsView creates a new view object attached the the global [gocui] screen object.
func newDetailsView(gui *gocui.Gui, efficiency float64, inefficiencies filetree.EfficiencySlice, imageSize uint64) (controller *Details) {
	controller = new(Details)

	// populate main fields
	controller.name = "details"
	controller.gui = gui
	controller.efficiency = efficiency
	controller.inefficiencies = inefficiencies
	controller.imageSize = imageSize

	return controller
}

func (v *Details) Name() string {
	return v.name
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (v *Details) Setup(view *gocui.View, header *gocui.View) error {
	logrus.Tracef("view.Setup() %s", v.Name())

	// set controller options
	v.view = view
	v.view.Editable = false
	v.view.Wrap = true
	v.view.Highlight = false
	v.view.Frame = false

	v.header = header
	v.header.Editable = false
	v.header.Wrap = false
	v.header.Frame = false

	var infos = []key.BindingInfo{
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
	}

	_, err := key.GenerateBindings(v.gui, v.name, infos)
	if err != nil {
		return err
	}

	return v.Render()
}

// IsVisible indicates if the details view pane is currently initialized.
func (v *Details) IsVisible() bool {
	return v != nil
}

// CursorDown moves the cursor down in the details pane (currently indicates nothing).
func (v *Details) CursorDown() error {
	return CursorDown(v.gui, v.view)
}

// CursorUp moves the cursor up in the details pane (currently indicates nothing).
func (v *Details) CursorUp() error {
	return CursorUp(v.gui, v.view)
}

// OnLayoutChange is called whenever the screen dimensions are changed
func (v *Details) OnLayoutChange() error {
	err := v.Update()
	if err != nil {
		return err
	}
	return v.Render()
}

// Update refreshes the state objects for future rendering.
func (v *Details) Update() error {
	return nil
}

func (v *Details) SetCurrentLayer(layer *image.Layer) {
	v.currentLayer = layer
}

// Render flushes the state objects to the screen. The details pane reports:
// 1. the current selected layer's command string
// 2. the image efficiency score
// 3. the estimated wasted image space
// 4. a list of inefficient file allocations
func (v *Details) Render() error {
	logrus.Tracef("view.Render() %s", v.Name())

	if v.currentLayer == nil {
		return fmt.Errorf("no layer selected")
	}

	var wastedSpace int64

	template := "%5s  %12s  %-s\n"
	inefficiencyReport := fmt.Sprintf(format.Header(template), "Count", "Total Space", "Path")

	height := 100
	if v.view != nil {
		_, height = v.view.Size()
	}

	for idx := 0; idx < len(v.inefficiencies); idx++ {
		data := v.inefficiencies[len(v.inefficiencies)-1-idx]
		wastedSpace += data.CumulativeSize

		// todo: make this report scrollable
		if idx < height {
			inefficiencyReport += fmt.Sprintf(template, strconv.Itoa(len(data.Nodes)), humanize.Bytes(uint64(data.CumulativeSize)), data.Path)
		}
	}

	imageSizeStr := fmt.Sprintf("%s %s", format.Header("Total Image size:"), humanize.Bytes(v.imageSize))
	effStr := fmt.Sprintf("%s %d %%", format.Header("Image efficiency score:"), int(100.0*v.efficiency))
	wastedSpaceStr := fmt.Sprintf("%s %s", format.Header("Potential wasted space:"), humanize.Bytes(uint64(wastedSpace)))

	v.gui.Update(func(g *gocui.Gui) error {
		// update header
		v.header.Clear()
		width, _ := v.view.Size()

		layerHeaderStr := format.RenderHeader("Layer Details", width, false)
		imageHeaderStr := format.RenderHeader("Image Details", width, false)

		_, err := fmt.Fprintln(v.header, layerHeaderStr)
		if err != nil {
			return err
		}

		// update contents
		v.view.Clear()

		var lines = make([]string, 0)
		if v.currentLayer.Names != nil && len(v.currentLayer.Names) > 0 {
			lines = append(lines, format.Header("Tags:   ")+strings.Join(v.currentLayer.Names, ", "))
		} else {
			lines = append(lines, format.Header("Tags:   ")+"(none)")
		}
		lines = append(lines, format.Header("Id:     ")+v.currentLayer.Id)
		lines = append(lines, format.Header("Digest: ")+v.currentLayer.Digest)
		lines = append(lines, format.Header("Command:"))
		lines = append(lines, v.currentLayer.Command)
		lines = append(lines, "\n"+imageHeaderStr)
		lines = append(lines, imageSizeStr)
		lines = append(lines, wastedSpaceStr)
		lines = append(lines, effStr+"\n")
		lines = append(lines, inefficiencyReport)

		_, err = fmt.Fprintln(v.view, strings.Join(lines, "\n"))
		if err != nil {
			logrus.Debug("unable to write to buffer: ", err)
		}
		return err
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected (currently does nothing).
func (v *Details) KeyHelp() string {
	return "TBD"
}
