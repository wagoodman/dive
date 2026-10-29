package export

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/internal/imagefixture"
)

// Test_ExportGoldens checks the JSON export of every image fixture, loaded through stereoscope.
// Regenerate with UPDATE_GOLDEN=true.
func Test_ExportGoldens(t *testing.T) {
	for _, name := range imagefixture.Names() {
		t.Run(name, func(t *testing.T) {
			analysis, err := dive.Analyze(context.Background(), name, imagefixture.Load(t, name))
			require.NoError(t, err)
			payload, err := NewExportFromAnalysis(analysis).Marshal()
			require.NoError(t, err)
			imagefixture.AssertGolden(t, imagefixture.GoldenPath(name, "export"), append(payload, '\n'))
		})
	}
}
