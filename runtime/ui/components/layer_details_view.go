package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wagoodman/dive/dive/image"
)

type LayerDetailModel interface {
	GetCurrentLayer() *image.Layer
}

// TODO make this scrollable
type LayerDetailsView struct {
	*tview.TextView
	LayerDetailModel
}

func NewLayerDetailsView(model LayerDetailModel) *LayerDetailsView {
	return &LayerDetailsView{
		TextView:         tview.NewTextView(),
		LayerDetailModel: model,
	}
}

func (lv *LayerDetailsView) Setup() *LayerDetailsView {
	lv.SetDynamicColors(true)
	return lv
}

func (lv *LayerDetailsView) getBox() *tview.Box {
	return lv.Box
}

func (lv *LayerDetailsView) getDraw() drawFn {
	return lv.Draw
}

func (lv *LayerDetailsView) getInputWrapper() inputFn {
	return lv.InputHandler
}

func (lv *LayerDetailsView) Draw(screen tcell.Screen) {
	displayText := layerDetailsText(lv.LayerDetailModel.GetCurrentLayer())
	lv.SetText(displayText)
	lv.TextView.Draw(screen)
}

func (lv *LayerDetailsView) GetKeyBindings() []KeyBindingDisplay {
	return []KeyBindingDisplay{}
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

