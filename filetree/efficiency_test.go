package filetree

import (
	"testing"
)

func TestEfficencyMap(t *testing.T) {
	trees := make([]*FileTree, 3)
	for idx := range trees {
		trees[idx] = NewFileTree()
	}

	trees[0].AddPath("/etc/nginx/nginx.conf", FileInfo{Size: 2000})
	trees[0].AddPath("/etc/nginx/public", FileInfo{Size: 3000})

	trees[1].AddPath("/etc/nginx/nginx.conf", FileInfo{Size: 5000})
	trees[1].AddPath("/etc/athing", FileInfo{Size: 10000})

	trees[2].AddPath("/etc/.wh.nginx", *BlankFileChangeInfo("/etc/.wh.nginx"))

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
		t.Fatalf("Expected to find %d inefficient path, but found %d", len(expectedMatches), len(actualMatches))
	}

	if expectedMatches[0].Path != actualMatches[0].Path {
		t.Errorf("Expected path of %s but go %s", expectedMatches[0].Path, actualMatches[0].Path)
	}

	if expectedMatches[0].CumulativeSize != actualMatches[0].CumulativeSize {
		t.Errorf("Expected cumulative size of %v but go %v", expectedMatches[0].CumulativeSize, actualMatches[0].CumulativeSize)
	}
}
