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

// DetailsView holds the UI objects and data models for populating the lower-left pane. Specifically the pane that
// shows the layer details and image statistics.
type DetailsView struct {
	Name           string
	gui            *gocui.Gui
	view           *gocui.View
	header         *gocui.View
	efficiency     float64
	inefficiencies filetree.EfficiencySlice
}

// NewDetailsView creates a new view object attached the the global [gocui] screen object.
func NewDetailsView(name string, gui *gocui.Gui, efficiency float64, inefficiencies filetree.EfficiencySlice) (detailsView *DetailsView) {
	detailsView = new(DetailsView)

	// populate main fields
	detailsView.Name = name
	detailsView.gui = gui
	detailsView.efficiency = efficiency
	detailsView.inefficiencies = inefficiencies

	return detailsView
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (view *DetailsView) Setup(v *gocui.View, header *gocui.View) error {

	// set view options
	view.view = v
	view.view.Editable = false
	view.view.Wrap = true
	view.view.Highlight = false
	view.view.Frame = false

	view.header = header
	view.header.Editable = false
	view.header.Wrap = false
	view.header.Frame = false

	// set keybindings
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowDown, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorDown() }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowUp, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorUp() }); err != nil {
		return err
	}

	return view.Render()
}

// IsVisible indicates if the details view pane is currently initialized.
func (view *DetailsView) IsVisible() bool {
	if view == nil {
		return false
	}
	return true
}

// CursorDown moves the cursor down in the details pane (currently indicates nothing).
func (view *DetailsView) CursorDown() error {
	return CursorDown(view.gui, view.view)
}

// CursorUp moves the cursor up in the details pane (currently indicates nothing).
func (view *DetailsView) CursorUp() error {
	return CursorUp(view.gui, view.view)
}

// Update refreshes the state objects for future rendering.
func (view *DetailsView) Update() error {
	return nil
}

// Render flushes the state objects to the screen. The details pane reports:
// 1. the current selected layer's command string
// 2. the image efficiency score
// 3. the estimated wasted image space
// 4. a list of inefficient file allocations
func (view *DetailsView) Render() error {
	currentLayer := Views.Layer.currentLayer()

	var wastedSpace int64

	template := "%5s  %12s  %-s\n"
	inefficiencyReport := fmt.Sprintf(Formatting.Header(template), "Count", "Total Space", "Path")

	height := 100
	if view.view != nil {
		_, height = view.view.Size()
	}

	for idx := 0; idx < len(view.inefficiencies); idx++ {
		data := view.inefficiencies[len(view.inefficiencies)-1-idx]
		wastedSpace += data.CumulativeSize

		// todo: make this report scrollable and exportable
		if idx < height {
			inefficiencyReport += fmt.Sprintf(template, strconv.Itoa(len(data.Nodes)), humanize.Bytes(uint64(data.CumulativeSize)), data.Path)
		}
	}

	imageSizeStr := fmt.Sprintf("%s %s", Formatting.Header("Total Image size:"), humanize.Bytes(Views.Layer.ImageSize))
	effStr := fmt.Sprintf("%s %d %%", Formatting.Header("Image efficiency score:"), int(100.0*view.efficiency))
	wastedSpaceStr := fmt.Sprintf("%s %s", Formatting.Header("Potential wasted space:"), humanize.Bytes(uint64(wastedSpace)))

	view.gui.Update(func(g *gocui.Gui) error {
		// update header
		view.header.Clear()
		width, _ := view.view.Size()

		layerHeaderStr := fmt.Sprintf("[Layer Details]%s", strings.Repeat("─", width-15))
		imageHeaderStr := fmt.Sprintf("[Image Details]%s", strings.Repeat("─", width-15))

		fmt.Fprintln(view.header, Formatting.Header(vtclean.Clean(layerHeaderStr, false)))

		// update contents
		view.view.Clear()
		fmt.Fprintln(view.view, Formatting.Header("Digest: ")+currentLayer.Id())
		// TODO: add back in with view model
		// fmt.Fprintln(view.view, Formatting.Header("Tar ID: ")+currentLayer.TarId())
		fmt.Fprintln(view.view, Formatting.Header("Command:"))
		fmt.Fprintln(view.view, currentLayer.Command())

		fmt.Fprintln(view.view, "\n"+Formatting.Header(vtclean.Clean(imageHeaderStr, false)))

		fmt.Fprintln(view.view, imageSizeStr)
		fmt.Fprintln(view.view, wastedSpaceStr)
		fmt.Fprintln(view.view, effStr+"\n")

		fmt.Fprintln(view.view, inefficiencyReport)
		return nil
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected (currently does nothing).
func (view *DetailsView) KeyHelp() string {
	return "TBD"
}
