package export

import (
	"testing"

	"github.com/wagoodman/dive/dive/image/docker"
)

func Test_Export(t *testing.T) {
	result := docker.TestAnalysisFromArchive(t, repoPath(t, ".data/test-docker-image.tar"))

	export := NewExport(result)
	payload, err := export.Marshal()
	if err != nil {
		t.Errorf("Test_Export: unable to export analysis: %v", err)
	}

	snaps.MatchJSON(t, payload)
}
