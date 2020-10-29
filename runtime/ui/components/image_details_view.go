package components

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/rivo/tview"
	"github.com/wagoodman/dive/dive/image"
	"strconv"
)


type ImageDetails struct {
	*tview.TextView
}

func NewImageDetailsView(analysisResult *image.AnalysisResult) *tview.TextView {
	result := tview.NewTextView()
	result.SetDynamicColors(true).
		SetScrollable(true)
	result.SetBorder(true).
		SetTitle("Image Details").
		SetTitleAlign(tview.AlignLeft)


	template := "%5s  %12s  %-s\n"
	inefficiencyReport := fmt.Sprintf(template, "[::b]Count[::-]", "[::b]Total Space[::-]", "[::b]Path[::-]")

	var wastedSpace int64 = 0
	height := 200

	for idx := 0; idx < len(analysisResult.Inefficiencies); idx++ {
		data := analysisResult.Inefficiencies[len(analysisResult.Inefficiencies)-1-idx]
		wastedSpace += data.CumulativeSize

		// todo: make this report scrollable
		if idx < height {
			inefficiencyReport += fmt.Sprintf(template, strconv.Itoa(len(data.Nodes)), humanize.Bytes(uint64(data.CumulativeSize)), data.Path)
		}
	}

	imageSizeStr := fmt.Sprintf("[::b]%s[::-] %s", "Total Image size:", humanize.Bytes(analysisResult.SizeBytes))
	effStr := fmt.Sprintf("[::b]%s[::-] %d %%", "Image efficiency score:", int(100.0*analysisResult.Efficiency))
	wastedSpaceStr := fmt.Sprintf("[::b]%s[::-] %s", "Potential wasted space:", humanize.Bytes(uint64(wastedSpace)))

	_, err := fmt.Fprintf(result,"%s\n%s\n%s\n%s", imageSizeStr, wastedSpaceStr, effStr, inefficiencyReport)
	if err != nil {
		panic(err)
	}

	return result
}