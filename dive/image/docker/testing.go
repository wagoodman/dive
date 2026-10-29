package docker

import (
	"fmt"
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

	result, err := image.Analyze(context.Background(), img)
	require.NoError(t, err, "unable to analyze image")
	return result
}

// TestAnalysisFromArchiveE is like TestAnalysisFromArchive but returns parse/analysis failures (including panics)
// instead of failing the test, so expected failures can be recorded.
func TestAnalysisFromArchiveE(t testing.TB, path string) (result *image.Analysis, err error) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	archive, err := TestLoadArchive(t, path)
	if err != nil {
		return nil, err
	}
	img, err := archive.ToImage(path)
	if err != nil {
		return nil, err
	}
	return image.Analyze(context.Background(), img)
}
