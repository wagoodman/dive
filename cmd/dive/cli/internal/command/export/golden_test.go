package export

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/dive/image/docker"
	"github.com/wagoodman/dive/internal/imagefixture"
)

// Test_ExportGoldens records the JSON export of every image fixture with the current (pre-stereoscope) parser, as the
// parity contract for the stereoscope implementation. Regenerate with UPDATE_GOLDEN=true.
func Test_ExportGoldens(t *testing.T) {
	for _, name := range imagefixture.Names() {
		t.Run(name, func(t *testing.T) {
			payload, err := exportFixture(t, imagefixture.Path(t, name))
			if err != nil {
				// the old parser can't read this fixture, record the failure as the expected outcome
				payload, err = json.MarshalIndent(map[string]string{"error": err.Error()}, "", "  ")
				require.NoError(t, err)
			}
			imagefixture.AssertGolden(t, imagefixture.GoldenPath(name, "export"), append(payload, '\n'))
		})
	}
}

func exportFixture(t *testing.T, path string) ([]byte, error) {
	analysis, err := docker.TestAnalysisFromArchiveE(t, path)
	if err != nil {
		return nil, err
	}
	return NewExport(analysis).Marshal()
}
