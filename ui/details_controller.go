package ui

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/jroimartin/gocui"
	"github.com/lunixbochs/vtclean"
	"github.com/wagoodman/dive/filetree"
	"strconv"
	"strings"
)

// DetailsController holds the UI objects and data models for populating the lower-left pane. Specifically the pane that
// shows the layer details and image statistics.
type DetailsController struct {
	Name           string
	gui            *gocui.Gui
	view           *gocui.View
	header         *gocui.View
	efficiency     float64
	inefficiencies filetree.EfficiencySlice
}

// NewDetailsController creates a new view object attached the the global [gocui] screen object.
func NewDetailsController(name string, gui *gocui.Gui, efficiency float64, inefficiencies filetree.EfficiencySlice) (controller *DetailsController) {
	controller = new(DetailsController)

	// populate main fields
	controller.Name = name
	controller.gui = gui
	controller.efficiency = efficiency
	controller.inefficiencies = inefficiencies

	return controller
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (controller *DetailsController) Setup(v *gocui.View, header *gocui.View) error {

	// set controller options
	controller.view = v
	controller.view.Editable = false
	controller.view.Wrap = true
	controller.view.Highlight = false
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

	return controller.Render()
}

// IsVisible indicates if the details view pane is currently initialized.
func (controller *DetailsController) IsVisible() bool {
	if controller == nil {
		return false
	}
	return true
}

// CursorDown moves the cursor down in the details pane (currently indicates nothing).
func (controller *DetailsController) CursorDown() error {
	return CursorDown(controller.gui, controller.view)
}

// CursorUp moves the cursor up in the details pane (currently indicates nothing).
func (controller *DetailsController) CursorUp() error {
	return CursorUp(controller.gui, controller.view)
}

// Update refreshes the state objects for future rendering.
func (controller *DetailsController) Update() error {
	return nil
}

// Render flushes the state objects to the screen. The details pane reports:
// 1. the current selected layer's command string
// 2. the image efficiency score
// 3. the estimated wasted image space
// 4. a list of inefficient file allocations
func (controller *DetailsController) Render() error {
	currentLayer := Controllers.Layer.currentLayer()

	var wastedSpace int64

	template := "%5s  %12s  %-s\n"
	inefficiencyReport := fmt.Sprintf(Formatting.Header(template), "Count", "Total Space", "Path")

	height := 100
	if controller.view != nil {
		_, height = controller.view.Size()
	}

	for idx := 0; idx < len(controller.inefficiencies); idx++ {
		data := controller.inefficiencies[len(controller.inefficiencies)-1-idx]
		wastedSpace += data.CumulativeSize

		// todo: make this report scrollable
		if idx < height {
			inefficiencyReport += fmt.Sprintf(template, strconv.Itoa(len(data.Nodes)), humanize.Bytes(uint64(data.CumulativeSize)), data.Path)
		}
	}

	imageSizeStr := fmt.Sprintf("%s %s", Formatting.Header("Total Image size:"), humanize.Bytes(Controllers.Layer.ImageSize))
	effStr := fmt.Sprintf("%s %d %%", Formatting.Header("Image efficiency score:"), int(100.0*controller.efficiency))
	wastedSpaceStr := fmt.Sprintf("%s %s", Formatting.Header("Potential wasted space:"), humanize.Bytes(uint64(wastedSpace)))

	controller.gui.Update(func(g *gocui.Gui) error {
		// update header
		controller.header.Clear()
		width, _ := controller.view.Size()

		layerHeaderStr := fmt.Sprintf("[Layer Details]%s", strings.Repeat("─", width-15))
		imageHeaderStr := fmt.Sprintf("[Image Details]%s", strings.Repeat("─", width-15))

		fmt.Fprintln(controller.header, Formatting.Header(vtclean.Clean(layerHeaderStr, false)))

		// update contents
		controller.view.Clear()
		fmt.Fprintln(controller.view, Formatting.Header("Digest: ")+currentLayer.Id())
		// TODO: add back in with controller model
		// fmt.Fprintln(view.view, Formatting.Header("Tar ID: ")+currentLayer.TarId())
		fmt.Fprintln(controller.view, Formatting.Header("Command:"))
		fmt.Fprintln(controller.view, currentLayer.Command())

		fmt.Fprintln(controller.view, "\n"+Formatting.Header(vtclean.Clean(imageHeaderStr, false)))

		fmt.Fprintln(controller.view, imageSizeStr)
		fmt.Fprintln(controller.view, wastedSpaceStr)
		fmt.Fprintln(controller.view, effStr+"\n")

		fmt.Fprintln(controller.view, inefficiencyReport)
		return nil
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected (currently does nothing).
func (controller *DetailsController) KeyHelp() string {
	return "TBD"
}
