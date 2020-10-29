package components

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wagoodman/dive/dive/image"
	"strings"
)

type LayerDetailModel interface {
	GetCurrentLayer() *image.Layer
}

// TODO make this scrollable
type LayerDetailsView struct {
	*tview.TextView
	layerIndex int
	LayerDetailModel
}

func NewLayerDetailsView(model LayerDetailModel) *LayerDetailsView {
	return &LayerDetailsView{
		TextView: tview.NewTextView(),
		LayerDetailModel: model,
	}
}

func (lv *LayerDetailsView) Setup() *LayerDetailsView {
	lv.SetTitle("Layer Details").SetTitleAlign(tview.AlignLeft)
	lv.SetDynamicColors(true).SetBorder(true)
	return lv
}

func (lv *LayerDetailsView) Draw(screen tcell.Screen) {
	displayText := layerDetailsText(lv.LayerDetailModel.GetCurrentLayer())
	lv.SetText(displayText)
	lv.TextView.Draw(screen)
}

func layerDetailsText(layer *image.Layer) string {
	lines := []string{}
	if layer.Names != nil && len(layer.Names) > 0 {
		lines = append(lines, boldString("Tags:   ")+strings.Join(layer.Names, ", "))
	} else {
		lines = append(lines, boldString("Tags:   ")+"(none)")
	}
	lines = append(lines, boldString("Id:     ")+layer.Id)
	lines = append(lines, boldString("Digest: ")+layer.Digest)
	lines = append(lines, boldString("Command:"))
	lines = append(lines, layer.Command)
	return strings.Join(lines, "\n")
}

func boldString(s string) string {
	return fmt.Sprintf("[::b]%s[::-]", s)
}