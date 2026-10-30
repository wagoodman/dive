package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalysis_PostLoad(t *testing.T) {
	tests := []struct {
		image, engine string
		wantEngine    string
	}{
		{image: "alpine", engine: "docker", wantEngine: "docker"},
		{image: "podman://alpine", engine: "docker", wantEngine: "docker"},
		{image: "img.tar", engine: "docker-tar", wantEngine: "docker-tar"},
		{image: "alpine", engine: "registry", wantEngine: "registry"},
		{image: "alpine", engine: "auto", wantEngine: "auto"},
		// invalid engines fall back to docker
		{image: "alpine", engine: "nope", wantEngine: "docker"},
	}
	for _, tt := range tests {
		t.Run(tt.image+" "+tt.engine, func(t *testing.T) {
			a := DefaultAnalysis()
			a.Image = tt.image
			a.Source = tt.engine
			require.NoError(t, a.PostLoad())
			// the reference is kept exactly as given, source.Parse splits it later
			assert.Equal(t, tt.image, a.Image)
			assert.Equal(t, tt.wantEngine, a.Source)
		})
	}
}

func TestAnalysis_PostLoad_platform(t *testing.T) {
	a := DefaultAnalysis()
	a.Platform = "linux/arm64"
	require.NoError(t, a.PostLoad())

	// caught before any image is loaded
	a.Platform = "a/b/c/d"
	require.ErrorContains(t, a.PostLoad(), "invalid platform")
}
