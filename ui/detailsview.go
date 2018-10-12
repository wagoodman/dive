package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/lunixbochs/vtclean"
	"strings"
	"github.com/wagoodman/dive/filetree"
	"strconv"
	"github.com/dustin/go-humanize"
)

type DetailsView struct {
	Name       string
	gui        *gocui.Gui
	view       *gocui.View
	header     *gocui.View
	efficiency float64
	inefficiencies filetree.EfficiencySlice
}

func NewStatisticsView(name string, gui *gocui.Gui) (detailsview *DetailsView) {
	detailsview = new(DetailsView)

	// populate main fields
	detailsview.Name = name
	detailsview.gui = gui

	return detailsview
}

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

func (view *DetailsView) IsVisible() bool {
	if view == nil {return false}
	return true
}

// we only need to update this view upon the initial tree load
func (view *DetailsView) Update() error {
	layerTrees := Views.Tree.RefTrees
	view.efficiency, view.inefficiencies = filetree.Efficiency(layerTrees[:len(layerTrees)-1])
	return nil
}

func (view *DetailsView) Render() error {
	currentLayer := Views.Layer.currentLayer()

	var wastedSpace int64

	template := "%5s  %12s  %-s\n"
	inefficiencyReport := fmt.Sprintf(Formatting.Header(template), "Count", "Total Space", "Path")
	for idx := len(view.inefficiencies)-1; idx > 0; idx-- {
		data := view.inefficiencies[idx]
		if data.CumulativeSize == 0 {
			continue
		}
		wastedSpace += data.CumulativeSize
		inefficiencyReport += fmt.Sprintf(template, strconv.Itoa(len(data.Nodes)), humanize.Bytes(uint64(data.CumulativeSize)), data.Path)
	}

	effStr := fmt.Sprintf("\n%s %d %%", Formatting.Header("Image efficiency score:"), int(100.0*view.efficiency))
	spaceStr := fmt.Sprintf("%s %s\n", Formatting.Header("Potential wasted space:"),  humanize.Bytes(uint64(wastedSpace)))

	view.gui.Update(func(g *gocui.Gui) error {
		// update header
		view.header.Clear()
		width, _ := g.Size()
		headerStr := fmt.Sprintf("[Image & Layer Details]%s", strings.Repeat("─",width*2))
		fmt.Fprintln(view.header, Formatting.Header(vtclean.Clean(headerStr, false)))

		// update contents
		view.view.Clear()
		fmt.Fprintln(view.view, Formatting.Header("Layer Command"))
		fmt.Fprintln(view.view, currentLayer.History.CreatedBy)

		fmt.Fprintln(view.view, effStr)
		fmt.Fprintln(view.view, spaceStr)

		fmt.Fprintln(view.view, inefficiencyReport)

		return nil
	})
	return nil
}

func (view *DetailsView) CursorDown() error {
	return CursorDown(view.gui, view.view)
}

func (view *DetailsView) CursorUp() error {
	return CursorUp(view.gui, view.view)
}


func (view *DetailsView) KeyHelp() string {
	return "TBD"
	// return  renderStatusOption("^L","Layer changes", view.CompareMode == CompareLayer) +
	// 		renderStatusOption("^A","All changes", view.CompareMode == CompareAll)
}
