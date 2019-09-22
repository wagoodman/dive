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
	analyzer := NewImageAnalyzer("dive-test:latest")
	err = analyzer.Parse(f)
	if err != nil {
		return nil, err
	}
	return analyzer.Analyze()
}
