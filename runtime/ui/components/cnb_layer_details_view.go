package components

import (
	"fmt"
	"strings"

	"github.com/buildpacks/lifecycle"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wagoodman/dive/dive/image"
)

type CNBLayerDetailModel interface {
	GetCurrentLayer() *image.Layer
	GetBOMFromDigest(layerSha string) lifecycle.BOMEntry
}

// TODO make this scrollable
type CNBLayerDetailsView struct {
	*tview.TextView
	layerIndex int
	CNBLayerDetailModel
}

func NewCNBLayerDetailsView(model CNBLayerDetailModel) *CNBLayerDetailsView {
	return &CNBLayerDetailsView{
		TextView:            tview.NewTextView(),
		CNBLayerDetailModel: model,
	}
}

func (lv *CNBLayerDetailsView) Setup() *CNBLayerDetailsView {
	lv.SetTitle("Layer Details").SetTitleAlign(tview.AlignLeft)
	lv.SetDynamicColors(true).SetBorder(true)
	return lv
}

func (lv *CNBLayerDetailsView) Draw(screen tcell.Screen) {
	curLayer := lv.CNBLayerDetailModel.GetCurrentLayer()
	curBOM := lv.GetBOMFromDigest(curLayer.Digest)
	displayText := layerCNBDetailsText(curLayer, curBOM)
	lv.SetText(displayText)
	lv.TextView.Draw(screen)
}


func (lv *CNBLayerDetailsView) GetKeyBindings() []KeyBindingDisplay {
	return []KeyBindingDisplay {}
}

func layerCNBDetailsText(layer *image.Layer, bom lifecycle.BOMEntry) string {
	lines := []string{}
	if layer.Names != nil && len(layer.Names) > 0 {
		lines = append(lines, boldString("Tags:   ")+strings.Join(layer.Names, ", "))
	} else {
		lines = append(lines, boldString("Tags:   ")+"(none)")
	}
	lines = append(lines, boldString("Id:     ")+layer.Id)
	lines = append(lines, layer.Command)
	lines = append(lines, boldString("BOM:   ")+fmt.Sprintf("Entry for: %s", bom.Buildpack.String()))
	return strings.Join(lines, "\n")
}

func (lv *CNBLayerDetailsView) getBox() *tview.Box {
	return lv.Box
}

func (lv *CNBLayerDetailsView) getDraw() drawFn {
	return lv.Draw
}

func (lv *CNBLayerDetailsView) getInputWrapper() inputFn {
	return lv.InputHandler
}
