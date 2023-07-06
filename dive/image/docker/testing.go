package docker

import (
	"os"
	"testing"

	"github.com/wagoodman/dive/dive/image"
)

func TestLoadArchive(tarPath string) (*ImageArchive, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return NewImageArchive(f)
}

func TestAnalysisFromArchive(t *testing.T, path string) *image.AnalysisResult {
	archive, err := TestLoadArchive(path)
	if err != nil {
		t.Fatalf("unable to fetch archive: %v", err)
	}

	img, err := archive.ToImage()
	if err != nil {
		t.Fatalf("unable to convert to image: %v", err)
	}

	result, err := img.Analyze()
	if err != nil {
		t.Fatalf("unable to analyze: %v", err)
	}
	return result
}
