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

// NOTE: this intentionally retains the original string concatenation behavior.
// A regression test will fail until this is optimized in a follow-up commit.
func renderImageDetailsBody(imageName string, imageSize uint64, efficiency float64, inefficiencies filetree.EfficiencySlice) (string, error) {
	inefficiencyReport := fmt.Sprintf(format.Header(imageDetailsAnalysisTemplate), "Count", "Total Space", "Path")

	var wastedSpace int64
	for idx := 0; idx < len(inefficiencies); idx++ {
		data := inefficiencies[len(inefficiencies)-1-idx]
		wastedSpace += data.CumulativeSize

		inefficiencyReport += fmt.Sprintf(
			imageDetailsAnalysisTemplate,
			strconv.Itoa(len(data.Nodes)),
			humanize.Bytes(uint64(data.CumulativeSize)),
			data.Path,
		)
	}

	imageNameStr := fmt.Sprintf("%s %s", format.Header("Image name:"), imageName)
	imageSizeStr := fmt.Sprintf("%s %s", format.Header("Total Image size:"), humanize.Bytes(imageSize))
	efficiencyStr := fmt.Sprintf("%s %d %%", format.Header("Image efficiency score:"), int(100.0*efficiency))
	wastedSpaceStr := fmt.Sprintf("%s %s", format.Header("Potential wasted space:"), humanize.Bytes(uint64(wastedSpace)))

	var lines = []string{
		imageNameStr,
		imageSizeStr,
		wastedSpaceStr,
		efficiencyStr,
		" ", // to avoid an empty line so CursorDown can work as expected
		inefficiencyReport,
	}

	return strings.Join(lines, "\n"), nil
}
