package ui

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/key"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/jroimartin/gocui"
	"github.com/lunixbochs/vtclean"
)

// detailsController holds the UI objects and data models for populating the lower-left pane. Specifically the pane that
// shows the layer details and image statistics.
type detailsController struct {
	name           string
	gui            *gocui.Gui
	view           *gocui.View
	header         *gocui.View
	efficiency     float64
	inefficiencies filetree.EfficiencySlice
}

// newDetailsController creates a new view object attached the the global [gocui] screen object.
func newDetailsController(name string, gui *gocui.Gui, efficiency float64, inefficiencies filetree.EfficiencySlice) (controller *detailsController) {
	controller = new(detailsController)

	// populate main fields
	controller.name = name
	controller.gui = gui
	controller.efficiency = efficiency
	controller.inefficiencies = inefficiencies

	return controller
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (controller *detailsController) Setup(v *gocui.View, header *gocui.View) error {

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

	var infos = []key.BindingInfo{
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
	}

	_, err := key.GenerateBindings(controller.gui, controller.name, infos)
	if err != nil {
		return err
	}

	return controller.Render()
}

// IsVisible indicates if the details view pane is currently initialized.
func (controller *detailsController) IsVisible() bool {
	return controller != nil
}

// CursorDown moves the cursor down in the details pane (currently indicates nothing).
func (controller *detailsController) CursorDown() error {
	return CursorDown(controller.gui, controller.view)
}

// CursorUp moves the cursor up in the details pane (currently indicates nothing).
func (controller *detailsController) CursorUp() error {
	return CursorUp(controller.gui, controller.view)
}

// Update refreshes the state objects for future rendering.
func (controller *detailsController) Update() error {
	return nil
}

// Render flushes the state objects to the screen. The details pane reports:
// 1. the current selected layer's command string
// 2. the image efficiency score
// 3. the estimated wasted image space
// 4. a list of inefficient file allocations
func (controller *detailsController) Render() error {
	currentLayer := controllers.Layer.currentLayer()

	var wastedSpace int64

	template := "%5s  %12s  %-s\n"
	inefficiencyReport := fmt.Sprintf(format.Header(template), "Count", "Total Space", "Path")

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

	imageSizeStr := fmt.Sprintf("%s %s", format.Header("Total Image size:"), humanize.Bytes(controllers.Layer.ImageSize))
	effStr := fmt.Sprintf("%s %d %%", format.Header("Image efficiency score:"), int(100.0*controller.efficiency))
	wastedSpaceStr := fmt.Sprintf("%s %s", format.Header("Potential wasted space:"), humanize.Bytes(uint64(wastedSpace)))

	controller.gui.Update(func(g *gocui.Gui) error {
		// update header
		controller.header.Clear()
		width, _ := controller.view.Size()

		layerHeaderStr := fmt.Sprintf("[Layer Details]%s", strings.Repeat("─", width-15))
		imageHeaderStr := fmt.Sprintf("[Image Details]%s", strings.Repeat("─", width-15))

		_, err := fmt.Fprintln(controller.header, format.Header(vtclean.Clean(layerHeaderStr, false)))
		if err != nil {
			return err
		}

		// update contents
		controller.view.Clear()

		var lines = make([]string, 0)
		if currentLayer.Names != nil && len(currentLayer.Names) > 0 {
			lines = append(lines, format.Header("Tags:   ")+strings.Join(currentLayer.Names, ", "))
		} else {
			lines = append(lines, format.Header("Tags:   ")+"(none)")
		}
		lines = append(lines, format.Header("Id:     ")+currentLayer.Id)
		lines = append(lines, format.Header("Digest: ")+currentLayer.Digest)
		lines = append(lines, format.Header("Command:"))
		lines = append(lines, currentLayer.Command)
		lines = append(lines, "\n"+format.Header(vtclean.Clean(imageHeaderStr, false)))
		lines = append(lines, imageSizeStr)
		lines = append(lines, wastedSpaceStr)
		lines = append(lines, effStr+"\n")
		lines = append(lines, inefficiencyReport)

		_, err = fmt.Fprintln(controller.view, strings.Join(lines, "\n"))
		if err != nil {
			logrus.Debug("unable to write to buffer: ", err)
		}
		return err
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected (currently does nothing).
func (controller *detailsController) KeyHelp() string {
	return "TBD"
}
