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

// A whiteout entry may reference a path that does not exist in any lower layer (for example, a file created and
// deleted within the same layer). That is not an error, and it must not abort the analysis of the layer: any
// duplicated files in the same layer still need to be reported.
func TestEfficiency_DanglingWhiteout(t *testing.T) {
	trees := make([]*FileTree, 2)
	for idx := range trees {
		trees[idx] = NewFileTree()
	}

	_, _, err := trees[0].AddPath("/etc/nginx/nginx.conf", FileInfo{Size: 100})
	checkError(t, err, "could not setup test")

	// re-adds the same path (wasting 100 bytes) and whites out a path that never existed below
	_, _, err = trees[1].AddPath("/etc/nginx/nginx.conf", FileInfo{Size: 100})
	checkError(t, err, "could not setup test")
	_, _, err = trees[1].AddPath("/etc/.wh.does-not-exist", *BlankFileChangeInfo("/etc/.wh.does-not-exist"))
	checkError(t, err, "could not setup test")

	actualScore, actualMatches := Efficiency(trees)

	if len(actualMatches) != 1 {
		for _, match := range actualMatches {
			t.Logf("   match: %+v", match)
		}
		t.Fatalf("expected 1 inefficient path, but found %d", len(actualMatches))
	}

	if actualMatches[0].Path != "/etc/nginx/nginx.conf" {
		t.Errorf("expected path of /etc/nginx/nginx.conf but got %s", actualMatches[0].Path)
	}

	if actualMatches[0].CumulativeSize != 200 {
		t.Errorf("expected cumulative size of 200 but got %v", actualMatches[0].CumulativeSize)
	}

	var expectedScore = 0.5
	if expectedScore != actualScore {
		t.Errorf("expected score of %v but got %v", expectedScore, actualScore)
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
