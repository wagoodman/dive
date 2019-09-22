package filetree

import (
	"testing"
)

func checkError(t *testing.T, err error, message string) {
	if err != nil {
		t.Errorf(message+": %+v", err)
	}
}

func TestEfficency(t *testing.T) {
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

func TestEfficency_ScratchImage(t *testing.T) {
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
