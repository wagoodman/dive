package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/dive"
)

func TestAnalysis_PostLoad(t *testing.T) {
	tests := []struct {
		image, engine string
		wantEngine    string
		wantImage     string
		wantSource    dive.ImageSource
	}{
		{image: "alpine", engine: "docker", wantEngine: "docker", wantImage: "alpine", wantSource: dive.SourceDockerEngine},
		{image: "podman://alpine", engine: "docker", wantEngine: "docker", wantImage: "alpine", wantSource: dive.SourcePodmanEngine},
		{image: "img.tar", engine: "docker-tar", wantEngine: "docker-tar", wantImage: "img.tar", wantSource: dive.SourceDockerArchive},
		// new sources are valid, the old path just doesn't know them
		{image: "alpine", engine: "registry", wantEngine: "registry", wantImage: "alpine", wantSource: dive.SourceUnknown},
		{image: "alpine", engine: "auto", wantEngine: "auto", wantImage: "alpine", wantSource: dive.SourceUnknown},
		{image: "oci-archive:img.tar", engine: "docker", wantEngine: "docker", wantImage: "oci-archive:img.tar", wantSource: dive.SourceDockerEngine},
		// invalid engines fall back to docker
		{image: "alpine", engine: "nope", wantEngine: "docker", wantImage: "alpine", wantSource: dive.SourceDockerEngine},
	}
	for _, tt := range tests {
		t.Run(tt.image+" "+tt.engine, func(t *testing.T) {
			a := DefaultAnalysis()
			a.Image = tt.image
			a.ContainerEngine = tt.engine
			require.NoError(t, a.PostLoad())
			assert.Equal(t, tt.image, a.Reference)
			assert.Equal(t, tt.wantEngine, a.ContainerEngine)
			assert.Equal(t, tt.wantImage, a.Image)
			assert.Equal(t, tt.wantSource, a.Source)
		})
	}
}
