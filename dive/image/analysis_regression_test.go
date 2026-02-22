package image

import (
	"context"
	"fmt"
	"testing"

	"github.com/wagoodman/dive/dive/filetree"
)

type regressionLayerFile struct {
	path string
	size int64
}

type regressionLayerSpec struct {
	size  uint64
	files []regressionLayerFile
}

func buildRegressionImage(t *testing.T, specs []regressionLayerSpec) *Image {
	t.Helper()

	trees := make([]*filetree.FileTree, len(specs))
	layers := make([]*Layer, len(specs))

	for idx, spec := range specs {
		tree := filetree.NewFileTree()
		for _, file := range spec.files {
			_, _, err := tree.AddPath(file.path, filetree.FileInfo{Size: file.size})
			if err != nil {
				t.Fatalf("could not add path %q in layer %d: %+v", file.path, idx, err)
			}
		}

		trees[idx] = tree
		layers[idx] = &Layer{
			Id:    fmt.Sprintf("layer-%d", idx),
			Index: idx,
			Size:  spec.size,
			Tree:  tree,
		}
	}

	return &Image{
		Request: "img:test",
		Trees:   trees,
		Layers:  layers,
	}
}

func TestAnalyzeWastedBytesCountsOnlyRemovableDuplicates(t *testing.T) {
	tests := map[string]struct {
		specs        []regressionLayerSpec
		wastedBytes  uint64
		wastedPct    float64
		userSizeByes uint64
	}{
		"single_copy_has_zero_waste": {
			specs: []regressionLayerSpec{
				{size: 100, files: []regressionLayerFile{{path: "/dup", size: 100}}},
				{size: 50, files: []regressionLayerFile{{path: "/other", size: 50}}},
			},
			wastedBytes:  0,
			wastedPct:    0,
			userSizeByes: 50,
		},
		"two_copies_remove_only_one": {
			specs: []regressionLayerSpec{
				{size: 100, files: []regressionLayerFile{{path: "/dup", size: 100}}},
				{size: 100, files: []regressionLayerFile{{path: "/dup", size: 100}}},
			},
			wastedBytes:  100,
			wastedPct:    1,
			userSizeByes: 100,
		},
		"ten_copies_remove_n_minus_one": {
			specs: func() []regressionLayerSpec {
				specs := make([]regressionLayerSpec, 10)
				for i := range specs {
					specs[i] = regressionLayerSpec{size: 100, files: []regressionLayerFile{{path: "/dup", size: 100}}}
				}
				return specs
			}(),
			wastedBytes:  900,
			wastedPct:    1,
			userSizeByes: 900,
		},
		"hundred_copies_remove_n_minus_one": {
			specs: func() []regressionLayerSpec {
				specs := make([]regressionLayerSpec, 100)
				for i := range specs {
					specs[i] = regressionLayerSpec{size: 10, files: []regressionLayerFile{{path: "/dup", size: 10}}}
				}
				return specs
			}(),
			wastedBytes:  990,
			wastedPct:    1,
			userSizeByes: 990,
		},
		"different_sizes_remove_all_but_smallest_copy": {
			specs: []regressionLayerSpec{
				{size: 400, files: []regressionLayerFile{{path: "/dup", size: 400}}},
				{size: 250, files: []regressionLayerFile{{path: "/dup", size: 250}}},
				{size: 100, files: []regressionLayerFile{{path: "/dup", size: 100}}},
				{size: 100, files: []regressionLayerFile{{path: "/dup", size: 100}}},
			},
			wastedBytes:  750,
			wastedPct:    750.0 / 450.0,
			userSizeByes: 450,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			analysis, err := Analyze(context.Background(), buildRegressionImage(t, test.specs))
			if err != nil {
				t.Fatalf("analyze failed: %+v", err)
			}

			if analysis.WastedBytes != test.wastedBytes {
				t.Fatalf("expected wasted bytes %d, got %d", test.wastedBytes, analysis.WastedBytes)
			}

			if analysis.UserSizeByes != test.userSizeByes {
				t.Fatalf("expected user size bytes %d, got %d", test.userSizeByes, analysis.UserSizeByes)
			}

			if analysis.WastedUserPercent != test.wastedPct {
				t.Fatalf("expected wasted percent %v, got %v", test.wastedPct, analysis.WastedUserPercent)
			}
		})
	}
}
