package view

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/format"
	"github.com/wagoodman/dive/dive/filetree"
)

const imageDetailsAnalysisTemplate = "%5s  %12s  %-s\n"

func renderImageDetailsBody(imageName string, imageSize uint64, efficiency float64, inefficiencies filetree.EfficiencySlice) (string, error) {
	var inefficiencyReport strings.Builder
	inefficiencyReport.Grow(len(inefficiencies) * 64)
	inefficiencyReport.WriteString(fmt.Sprintf(format.Header(imageDetailsAnalysisTemplate), "Count", "Total Space", "Path"))

	var wastedSpace int64
	for idx := 0; idx < len(inefficiencies); idx++ {
		data := inefficiencies[len(inefficiencies)-1-idx]
		wastedSpace += data.CumulativeSize

		_, err := fmt.Fprintf(
			&inefficiencyReport,
			imageDetailsAnalysisTemplate,
			strconv.Itoa(len(data.Nodes)),
			humanize.Bytes(uint64(data.CumulativeSize)),
			data.Path,
		)
		if err != nil {
			return "", err
		}
	}

	var body strings.Builder
	imageNameStr := fmt.Sprintf("%s %s", format.Header("Image name:"), imageName)
	imageSizeStr := fmt.Sprintf("%s %s", format.Header("Total Image size:"), humanize.Bytes(imageSize))
	efficiencyStr := fmt.Sprintf("%s %d %%", format.Header("Image efficiency score:"), int(100.0*efficiency))
	wastedSpaceStr := fmt.Sprintf("%s %s", format.Header("Potential wasted space:"), humanize.Bytes(uint64(wastedSpace)))

	body.WriteString(imageNameStr)
	body.WriteString("\n")
	body.WriteString(imageSizeStr)
	body.WriteString("\n")
	body.WriteString(wastedSpaceStr)
	body.WriteString("\n")
	body.WriteString(efficiencyStr)
	body.WriteString("\n")
	body.WriteString(" ")
	body.WriteString("\n")
	body.WriteString(inefficiencyReport.String())

	return body.String(), nil
}
