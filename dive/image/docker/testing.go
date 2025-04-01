package docker

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"os"
	"testing"

	"github.com/wagoodman/dive/dive/image"
)

func TestLoadArchive(t testing.TB, tarPath string) (*ImageArchive, error) {
	t.Helper()
	f, err := os.Open(tarPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return NewImageArchive(f)
}

func TestAnalysisFromArchive(t testing.TB, path string) *image.Analysis {
	t.Helper()
	archive, err := TestLoadArchive(t, path)
	require.NoError(t, err, "unable to load archive")

	img, err := archive.ToImage(path)
	require.NoError(t, err, "unable to convert archive to image")

	result, err := img.Analyze(context.Background())
	require.NoError(t, err, "unable to analyze image")
	return result
}
