package filetree

import (
	"testing"
)

func checkError(t *testing.T, err error, message string) {
	if err != nil {
		t.Errorf(message+": %+v", err)
	}
}

func TestEfficiency(t *testing.T) {
	trees := make([]*FileTree, 3)
	for idx := range trees {
		trees[idx] = NewFileTree()
	}

	_, _, err := trees[0].AddPath("/etc/nginx/nginx.conf", FileInfo{Size: 2000})
	checkError(t, err, "could not setup test")

	_, _, err = trees[0].AddPath("/etc/nginx/public", FileInfo{Size: 3000})
	checkError(t, err, "could not setup test")

	_, _, err = trees[1].AddPath("/etc/nginx/nginx.conf", FileInfo{Size: 5000})
	checkError(t, err, "could not setup test")
	_, _, err = trees[1].AddPath("/etc/athing", FileInfo{Size: 10000})
	checkError(t, err, "could not setup test")

	_, _, err = trees[2].AddPath("/etc/.wh.nginx", *BlankFileChangeInfo("/etc/.wh.nginx"))
	checkError(t, err, "could not setup test")

	var expectedScore = 0.75
	var expectedMatches = EfficiencySlice{
		&EfficiencyData{Path: "/etc/nginx/nginx.conf", CumulativeSize: 7000},
	}
	actualScore, actualMatches := Efficiency(trees)

	if expectedScore != actualScore {
		t.Errorf("Expected score of %v but go %v", expectedScore, actualScore)
	}

	if len(actualMatches) != len(expectedMatches) {
		for _, match := range actualMatches {
			t.Logf("   match: %+v", match)
		}
		t.Fatalf("Expected to find %d inefficient paths, but found %d", len(expectedMatches), len(actualMatches))
	}

	if expectedMatches[0].Path != actualMatches[0].Path {
		t.Errorf("Expected path of %s but go %s", expectedMatches[0].Path, actualMatches[0].Path)
	}

	if expectedMatches[0].CumulativeSize != actualMatches[0].CumulativeSize {
		t.Errorf("Expected cumulative size of %v but go %v", expectedMatches[0].CumulativeSize, actualMatches[0].CumulativeSize)
	}
}

func TestEfficiency_ScratchImage(t *testing.T) {
	trees := make([]*FileTree, 3)
	for idx := range trees {
		trees[idx] = NewFileTree()
	}

	_, _, err := trees[0].AddPath("/nothing", FileInfo{Size: 0})
	checkError(t, err, "could not setup test")

	var expectedScore = 1.0
	var expectedMatches = EfficiencySlice{}
	actualScore, actualMatches := Efficiency(trees)

	if expectedScore != actualScore {
		t.Errorf("Expected score of %v but go %v", expectedScore, actualScore)
	}

	if len(actualMatches) > 0 {
		t.Fatalf("Expected to find %d inefficient paths, but found %d", len(expectedMatches), len(actualMatches))
	}

}

func TestEfficiency_ManyCopiesAndHashSizeVariation(t *testing.T) {
	tests := map[string]struct {
		path      string
		sizes     []int64
		hashes    []uint64
		expected  int64
		expectedN int
	}{
		"hundred_copies": {
			path: "/dup",
			sizes: func() []int64 {
				vals := make([]int64, 100)
				for i := range vals {
					vals[i] = 10
				}
				return vals
			}(),
			hashes: func() []uint64 {
				vals := make([]uint64, 100)
				for i := range vals {
					vals[i] = uint64(i + 1)
				}
				return vals
			}(),
			expected:  1000,
			expectedN: 100,
		},
		"different_hashes_same_size": {
			path:      "/same-size",
			sizes:     []int64{100, 100, 100},
			hashes:    []uint64{1, 2, 3},
			expected:  300,
			expectedN: 3,
		},
		"different_sizes_and_hashes": {
			path:      "/varying",
			sizes:     []int64{400, 250, 100, 100},
			hashes:    []uint64{10, 20, 30, 40},
			expected:  850,
			expectedN: 4,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			trees := make([]*FileTree, len(test.sizes))
			for idx := range test.sizes {
				trees[idx] = NewFileTree()
				_, _, err := trees[idx].AddPath(test.path, FileInfo{Size: test.sizes[idx], hash: test.hashes[idx]})
				if err != nil {
					t.Fatalf("could not setup test at layer %d: %+v", idx, err)
				}
			}

			_, matches := Efficiency(trees)

			if len(matches) != 1 {
				t.Fatalf("expected 1 inefficiency match, got %d", len(matches))
			}

			if matches[0].CumulativeSize != test.expected {
				t.Fatalf("expected cumulative size %d, got %d", test.expected, matches[0].CumulativeSize)
			}

			if len(matches[0].Nodes) != test.expectedN {
				t.Fatalf("expected node count %d, got %d", test.expectedN, len(matches[0].Nodes))
			}
		})
	}
}
