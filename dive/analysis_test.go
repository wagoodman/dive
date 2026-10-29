package dive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/internal/imagefixture"
)

func TestAnalyze_Efficiency(t *testing.T) {
	tests := []struct {
		fixture        string
		efficiency     float64
		inefficiencies []Inefficiency
	}{
		{
			// every copy counts toward the waste, the final whiteout of big.bin adds an occurrence of 0 bytes,
			// and equal sizes are ordered by path
			fixture:    "duplicate-files-across-layers",
			efficiency: 4000.0 / 38000.0,
			inefficiencies: []Inefficiency{
				{Path: "/var/cache/big.bin", Count: 4, SizeBytes: 30000},
				{Path: "/var/cache/medium.bin", Count: 2, SizeBytes: 6000},
				{Path: "/var/cache/tie-a", Count: 2, SizeBytes: 1000},
				{Path: "/var/cache/tie-b", Count: 2, SizeBytes: 1000},
			},
		},
		{
			// a removed file is wasted at its original size. A removed dir is one occurrence sized as everything
			// under it (x + nested/y = 1100), which lowers the score but is never listed as waste
			fixture:    "file-and-dir-whiteouts",
			efficiency: 2275.0 / 2675.0,
			inefficiencies: []Inefficiency{
				{Path: "/data/remove.txt", Count: 2, SizeBytes: 400},
			},
		},
		{
			// opaque whiteouts are ignored, so the hidden lower files are not waste
			fixture:    "opaque-whiteout-dir",
			efficiency: 1,
		},
		{
			// a whiteout of a path no lower layer has is skipped, the rest of the layer still counts
			fixture:    "whiteout-missing-target",
			efficiency: 150.0 / 300.0,
			inefficiencies: []Inefficiency{
				{Path: "/opt/a", Count: 2, SizeBytes: 250},
			},
		},
		{
			// hardlinks are 0 bytes, as in the tar header
			fixture:    "hardlink-and-symlink",
			efficiency: 1400.0 / 1411.0,
			inefficiencies: []Inefficiency{
				{Path: "/usr/bin/tool-rel", Count: 2, SizeBytes: 11},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.fixture, func(t *testing.T) {
			analysis, err := Analyze(context.Background(), tt.fixture, imagefixture.Load(t, tt.fixture))
			require.NoError(t, err)

			assert.InDelta(t, tt.efficiency, analysis.Efficiency, 1e-12)
			assert.Equal(t, tt.inefficiencies, analysis.Inefficiencies)

			var wasted uint64
			for _, ineff := range tt.inefficiencies {
				wasted += ineff.SizeBytes
			}
			assert.Equal(t, wasted, analysis.WastedBytes)
		})
	}
}
