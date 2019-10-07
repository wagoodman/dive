package docker

import (
	"github.com/wagoodman/dive/dive/image"
	"os"
)

func TestLoadDockerImageTar(tarPath string) (*image.AnalysisResult, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	archive, err := NewImageArchive(f)
	if err != nil {
		return nil, err
	}

	img, err := archive.ToImage()
	if err != nil {
		return nil, err
	}

	return img.Analyze()
}
