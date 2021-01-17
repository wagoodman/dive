package components

import (
	"fmt"
	"strconv"

	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/format"
)

type ImageDetails struct {
	*tview.TextView
	analysisResult *image.AnalysisResult
}

func NewImageDetailsView(analysisResult *image.AnalysisResult) *ImageDetails {
	return &ImageDetails{
		TextView:       tview.NewTextView(),
		analysisResult: analysisResult,
	}
}

func (lv *ImageDetails) Setup() *ImageDetails {
	lv.SetDynamicColors(true).SetScrollable(true)
	return lv
}

func (lv *ImageDetails) getBox() *tview.Box {
	return lv.Box
}

func (lv *ImageDetails) getDraw() drawFn {
	return lv.Draw
}

func (lv *ImageDetails) getInputWrapper() inputFn {
	return lv.InputHandler
}

func (lv *ImageDetails) Draw(screen tcell.Screen) {
	lv.SetText(lv.imageDetailsText())
	lv.TextView.Draw(screen)
}

func (lv *ImageDetails) imageDetailsText() string {
	template := format.Header("%5s  %12s  %s\n")
	inefficiencyReport := fmt.Sprintf(template, "Count", "Total Space", "Path")

	var wastedSpace int64 = 0
	height := 200

	for idx := 0; idx < len(lv.analysisResult.Inefficiencies); idx++ {
		data := lv.analysisResult.Inefficiencies[len(lv.analysisResult.Inefficiencies)-1-idx]
		wastedSpace += data.CumulativeSize

		// todo: make this report scrollable
		if idx < height {
			inefficiencyReport += fmt.Sprintf(template, strconv.Itoa(len(data.Nodes)), humanize.Bytes(uint64(data.CumulativeSize)), data.Path)
		}
	}

	imageSizeStr := fmt.Sprintf("%s %s", format.Header("Total Image size:"), humanize.Bytes(lv.analysisResult.SizeBytes))
	effStr := fmt.Sprintf("%s %d %%", format.Header("Image efficiency score:"), int(100.0*lv.analysisResult.Efficiency))
	wastedSpaceStr := fmt.Sprintf("%s %s", format.Header("Potential wasted space:"), humanize.Bytes(uint64(wastedSpace)))

	return fmt.Sprintf("%s\n%s\n%s\n%s", imageSizeStr, wastedSpaceStr, effStr, inefficiencyReport)
}

func (lv *ImageDetails) GetKeyBindings() []KeyBindingDisplay {
	return []KeyBindingDisplay{}
}
