package dive

import (
	"archive/tar"
	"context"
	"strings"
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
			// every copy but the one left in the image is waste, all of big.bin since it ends up removed, and
			// equal sizes are ordered by path
			fixture:    "duplicate-files-across-layers",
			efficiency: 4000.0 / 38000.0,
			inefficiencies: []Inefficiency{
				{Path: "/var/cache/big.bin", Count: 4, SizeBytes: 30000},
				{Path: "/var/cache/medium.bin", Count: 2, SizeBytes: 3000},
				{Path: "/var/cache/tie-a", Count: 2, SizeBytes: 500},
				{Path: "/var/cache/tie-b", Count: 2, SizeBytes: 500},
			},
		},
		{
			// a removed file is wasted at its original size, and so is every file under a removed dir, even when
			// the dir is made again later
			fixture:    "file-and-dir-whiteouts",
			efficiency: 75.0 / 1575.0,
			inefficiencies: []Inefficiency{
				{Path: "/data/olddir/nested/y", Count: 2, SizeBytes: 600},
				{Path: "/data/olddir/x", Count: 2, SizeBytes: 500},
				{Path: "/data/remove.txt", Count: 2, SizeBytes: 400},
			},
		},
		{
			// an opaque dir removes the lower files it hides
			fixture:    "opaque-whiteout-dir",
			efficiency: 55.0 / 655.0,
			inefficiencies: []Inefficiency{
				{Path: "/etc/app/sub/c.conf", Count: 2, SizeBytes: 300},
				{Path: "/etc/app/b.conf", Count: 2, SizeBytes: 200},
				{Path: "/etc/app/a.conf", Count: 2, SizeBytes: 100},
			},
		},
		{
			// a whiteout of a path no lower layer has is skipped, the rest of the layer still counts
			fixture:    "whiteout-missing-target",
			efficiency: 200.0 / 300.0,
			inefficiencies: []Inefficiency{
				{Path: "/opt/a", Count: 2, SizeBytes: 100},
			},
		},
		{
			// hardlinks are 0 bytes, as in the tar header, and a file over a symlink leaves no bytes behind
			fixture:    "hardlink-and-symlink",
			efficiency: 1,
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

func TestAnalyze_WasteIsWhatTheFinalImageCannotReach(t *testing.T) {
	data := func(content string) []imagefixture.Entry {
		return []imagefixture.Entry{{Name: "data", Content: content}}
	}
	tests := []struct {
		name          string
		layers        [][]imagefixture.Entry
		wasted        uint64
		efficiency    float64
		wastedPercent float64
	}{
		{
			name:          "rewritten unchanged",
			layers:        [][]imagefixture.Entry{data(strings.Repeat("a", 1000)), data(strings.Repeat("a", 1000))},
			wasted:        1000,
			efficiency:    0.5,
			wastedPercent: 1,
		},
		{
			name:          "grown",
			layers:        [][]imagefixture.Entry{data(strings.Repeat("a", 10)), data(strings.Repeat("a", 100))},
			wasted:        10,
			efficiency:    100.0 / 110.0,
			wastedPercent: 0.1,
		},
		{
			name: "removed then added back",
			layers: [][]imagefixture.Entry{
				data(strings.Repeat("a", 1000)),
				{{Name: ".wh.data"}},
				data(strings.Repeat("b", 1000)),
			},
			wasted:        1000,
			efficiency:    0.5,
			wastedPercent: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := imagefixture.LoadSpec(t, imagefixture.Spec{Format: imagefixture.OCISave, Layers: tt.layers})
			analysis, err := Analyze(context.Background(), "spec", img)
			require.NoError(t, err)

			require.Len(t, analysis.Inefficiencies, 1)
			assert.Equal(t, tt.wasted, analysis.Inefficiencies[0].SizeBytes)
			assert.Equal(t, tt.wasted, analysis.WastedBytes)
			assert.InDelta(t, tt.efficiency, analysis.Efficiency, 1e-12)
			assert.InDelta(t, tt.wastedPercent, analysis.WastedUserPercent, 1e-12)
		})
	}
}

func TestAnalyze_EmptyPathsAreNotWaste(t *testing.T) {
	// empty files and dirs repeated in every layer have no bytes to waste
	layer := []imagefixture.Entry{
		{Name: "etc/", Type: tar.TypeDir},
		{Name: "etc/empty.conf"},
		{Name: "var/empty/", Type: tar.TypeDir},
	}
	img := imagefixture.LoadSpec(t, imagefixture.Spec{
		Format: imagefixture.OCISave,
		Layers: [][]imagefixture.Entry{layer, layer, layer},
	})
	analysis, err := Analyze(context.Background(), "spec", img)
	require.NoError(t, err)
	assert.Empty(t, analysis.Inefficiencies)
	assert.Zero(t, analysis.WastedBytes)
	assert.Equal(t, 1.0, analysis.Efficiency)
}

func TestAnalyze_RemovedDirs(t *testing.T) {
	base := []imagefixture.Entry{
		{Name: "data/", Type: tar.TypeDir},
		{Name: "data/tmp/", Type: tar.TypeDir},
		{Name: "data/tmp/a", Content: strings.Repeat("a", 1000)},
		{Name: "data/tmp/b", Content: strings.Repeat("b", 1000)},
	}
	tests := []struct {
		name  string
		upper []imagefixture.Entry
		// wasted maps each removed path to its wasted bytes
		wasted map[string]uint64
	}{
		{
			name:   "dir whiteout",
			upper:  []imagefixture.Entry{{Name: "data/", Type: tar.TypeDir}, {Name: "data/.wh.tmp"}},
			wasted: map[string]uint64{"/data/tmp/a": 1000, "/data/tmp/b": 1000},
		},
		{
			name: "opaque dir keeps what the layer adds back",
			upper: []imagefixture.Entry{
				{Name: "data/", Type: tar.TypeDir},
				{Name: "data/tmp/", Type: tar.TypeDir},
				{Name: "data/tmp/.wh..wh..opq"},
				{Name: "data/tmp/a", Content: "A"},
			},
			wasted: map[string]uint64{"/data/tmp/a": 1000, "/data/tmp/b": 1000},
		},
		{
			name: "dir replaced by a symlink",
			upper: []imagefixture.Entry{
				{Name: "data/", Type: tar.TypeDir},
				{Name: "data/tmp", Type: tar.TypeSymlink, Linkname: "/tmp"},
			},
			wasted: map[string]uint64{"/data/tmp/a": 1000, "/data/tmp/b": 1000},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := imagefixture.LoadSpec(t, imagefixture.Spec{
				Format: imagefixture.OCISave,
				Layers: [][]imagefixture.Entry{base, tt.upper},
			})
			analysis, err := Analyze(context.Background(), "spec", img)
			require.NoError(t, err)

			got := map[string]uint64{}
			for _, ineff := range analysis.Inefficiencies {
				got[ineff.Path] = ineff.SizeBytes
			}
			assert.Equal(t, tt.wasted, got)
			assert.Less(t, analysis.Efficiency, 1.0)
		})
	}
}

func TestAnalyze_SingleLayer(t *testing.T) {
	// the only layer is the base, so there are no user bytes to be wasted
	analysis, err := Analyze(context.Background(), "oci-archive-gzip", imagefixture.Load(t, "oci-archive-gzip"))
	require.NoError(t, err)
	require.Len(t, analysis.Layers, 1)
	assert.Zero(t, analysis.UserSizeBytes)
	assert.Zero(t, analysis.WastedUserPercent)
}

func TestAnalyze_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Analyze(ctx, "legacy-save", imagefixture.Load(t, "legacy-save"))
	assert.ErrorIs(t, err, context.Canceled)
}
