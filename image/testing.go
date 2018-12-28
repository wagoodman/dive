package image

import (
	"os"
)

func TestLoadDockerImageTar(tarPath string) (*AnalysisResult, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	analyzer := newDockerImageAnalyzer("dive-test:latest")
	err = analyzer.Parse(f)
	if err != nil {
		return nil, err
	}
	return analyzer.Analyze()
}
