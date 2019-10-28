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
	"github.com/lunixbochs/vtclean"
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

// NewDetailsView creates a new view object attached the the global [gocui] screen object.
func NewDetailsView(name string, gui *gocui.Gui, efficiency float64, inefficiencies filetree.EfficiencySlice, imageSize uint64) (controller *Details) {
	controller = new(Details)

	// populate main fields
	controller.name = name
	controller.gui = gui
	controller.efficiency = efficiency
	controller.inefficiencies = inefficiencies
	controller.imageSize = imageSize

	return controller
}

func (c *Details) Name() string {
	return c.name
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (c *Details) Setup(v *gocui.View, header *gocui.View) error {

	// set controller options
	c.view = v
	c.view.Editable = false
	c.view.Wrap = true
	c.view.Highlight = false
	c.view.Frame = false

	c.header = header
	c.header.Editable = false
	c.header.Wrap = false
	c.header.Frame = false

	var infos = []key.BindingInfo{
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
	}

	_, err := key.GenerateBindings(c.gui, c.name, infos)
	if err != nil {
		return err
	}

	return c.Render()
}

// IsVisible indicates if the details view pane is currently initialized.
func (c *Details) IsVisible() bool {
	return c != nil
}

// CursorDown moves the cursor down in the details pane (currently indicates nothing).
func (c *Details) CursorDown() error {
	return CursorDown(c.gui, c.view)
}

// CursorUp moves the cursor up in the details pane (currently indicates nothing).
func (c *Details) CursorUp() error {
	return CursorUp(c.gui, c.view)
}

// Update refreshes the state objects for future rendering.
func (c *Details) Update() error {
	return nil
}

func (c *Details) SetCurrentLayer(layer *image.Layer) {
	c.currentLayer = layer
}

// Render flushes the state objects to the screen. The details pane reports:
// 1. the current selected layer's command string
// 2. the image efficiency score
// 3. the estimated wasted image space
// 4. a list of inefficient file allocations
func (c *Details) Render() error {
	if c.currentLayer == nil {
		return fmt.Errorf("no layer selected")
	}

	var wastedSpace int64

	template := "%5s  %12s  %-s\n"
	inefficiencyReport := fmt.Sprintf(format.Header(template), "Count", "Total Space", "Path")

	height := 100
	if c.view != nil {
		_, height = c.view.Size()
	}

	for idx := 0; idx < len(c.inefficiencies); idx++ {
		data := c.inefficiencies[len(c.inefficiencies)-1-idx]
		wastedSpace += data.CumulativeSize

		// todo: make this report scrollable
		if idx < height {
			inefficiencyReport += fmt.Sprintf(template, strconv.Itoa(len(data.Nodes)), humanize.Bytes(uint64(data.CumulativeSize)), data.Path)
		}
	}

	imageSizeStr := fmt.Sprintf("%s %s", format.Header("Total Image size:"), humanize.Bytes(c.imageSize))
	effStr := fmt.Sprintf("%s %d %%", format.Header("Image efficiency score:"), int(100.0*c.efficiency))
	wastedSpaceStr := fmt.Sprintf("%s %s", format.Header("Potential wasted space:"), humanize.Bytes(uint64(wastedSpace)))

	c.gui.Update(func(g *gocui.Gui) error {
		// update header
		c.header.Clear()
		width, _ := c.view.Size()

		layerHeaderStr := fmt.Sprintf("[Layer Details]%s", strings.Repeat("─", width-15))
		imageHeaderStr := fmt.Sprintf("[Image Details]%s", strings.Repeat("─", width-15))

		_, err := fmt.Fprintln(c.header, format.Header(vtclean.Clean(layerHeaderStr, false)))
		if err != nil {
			return err
		}

		// update contents
		c.view.Clear()

		var lines = make([]string, 0)
		if c.currentLayer.Names != nil && len(c.currentLayer.Names) > 0 {
			lines = append(lines, format.Header("Tags:   ")+strings.Join(c.currentLayer.Names, ", "))
		} else {
			lines = append(lines, format.Header("Tags:   ")+"(none)")
		}
		lines = append(lines, format.Header("Id:     ")+c.currentLayer.Id)
		lines = append(lines, format.Header("Digest: ")+c.currentLayer.Digest)
		lines = append(lines, format.Header("Command:"))
		lines = append(lines, c.currentLayer.Command)
		lines = append(lines, "\n"+format.Header(vtclean.Clean(imageHeaderStr, false)))
		lines = append(lines, imageSizeStr)
		lines = append(lines, wastedSpaceStr)
		lines = append(lines, effStr+"\n")
		lines = append(lines, inefficiencyReport)

		_, err = fmt.Fprintln(c.view, strings.Join(lines, "\n"))
		if err != nil {
			logrus.Debug("unable to write to buffer: ", err)
		}
		return err
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected (currently does nothing).
func (c *Details) KeyHelp() string {
	return "TBD"
}
