package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/dive"
)

func Test_Analysis_ContainerEngine_FromConfig_Podman(t *testing.T) {
	// container-engine: podman in config (current key) should resolve to SourcePodmanEngine.
	a := DefaultAnalysis()
	a.ContainerEngine = "podman"
	a.Image = "myimage:latest"

	require.NoError(t, a.PostLoad())
	assert.Equal(t, dive.SourcePodmanEngine, a.Source,
		"expected podman source when container-engine=podman and image has no scheme prefix")
}

func Test_Analysis_ContainerEngine_FromConfig_Docker(t *testing.T) {
	// Default: docker engine
	a := DefaultAnalysis()
	a.Image = "myimage:latest"

	require.NoError(t, a.PostLoad())
	assert.Equal(t, dive.SourceDockerEngine, a.Source,
		"expected docker source when container-engine=docker (default)")
}

func Test_Analysis_ContainerEngine_SchemeOverridesConfig(t *testing.T) {
	// An explicit scheme prefix on the image overrides container-engine config.
	a := DefaultAnalysis()
	a.ContainerEngine = "podman"
	a.Image = "docker://myimage:latest"

	require.NoError(t, a.PostLoad())
	assert.Equal(t, dive.SourceDockerEngine, a.Source,
		"expected docker source when image has docker:// prefix, regardless of container-engine config")
}

func Test_Analysis_LegacySourceKey_Podman(t *testing.T) {
	// Legacy "source" config key (used before v0.14 CLI refactor) should still work when
	// "container-engine" is at its default value, and should emit a deprecation warning.
	a := DefaultAnalysis()
	a.LegacySource = "podman"
	a.Image = "myimage:latest"

	require.NoError(t, a.PostLoad())
	assert.Equal(t, dive.SourcePodmanEngine, a.Source,
		"expected podman source when legacy source=podman and container-engine is at default")
}

func Test_Analysis_LegacySourceKey_DoesNotOverrideExplicitContainerEngine(t *testing.T) {
	// If the user has both keys, "container-engine" wins over the legacy "source".
	a := DefaultAnalysis()
	a.ContainerEngine = "podman"
	a.LegacySource = "docker"
	a.Image = "myimage:latest"

	require.NoError(t, a.PostLoad())
	assert.Equal(t, dive.SourcePodmanEngine, a.Source,
		"container-engine should take precedence over legacy source key")
}
