package export

import (
	"context"
	"testing"

	"github.com/anchore/stereoscope/pkg/image"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/internal/imagefixture"
	"github.com/wagoodman/dive/internal/source"
)

// loadFixture reads the fixture through dive's own source resolution, the way the CLI does
func loadFixture(t *testing.T, name string) *image.Image {
	t.Helper()
	img, err := source.Load(context.Background(), imagefixture.Reference(t, name), source.Config{})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, img.Cleanup()) })
	return img
}

// Test_ExportGoldens checks the JSON export of every image fixture.
// Regenerate with UPDATE_GOLDEN=true.
func Test_ExportGoldens(t *testing.T) {
	for _, name := range imagefixture.Names() {
		t.Run(name, func(t *testing.T) {
			analysis, err := dive.Analyze(context.Background(), name, loadFixture(t, name))
			require.NoError(t, err)
			payload, err := NewExport(analysis).Marshal()
			require.NoError(t, err)
			imagefixture.AssertGolden(t, imagefixture.GoldenPath(name, "export"), append(payload, '\n'))
		})
	}
}
