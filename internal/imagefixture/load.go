package imagefixture

import (
	"context"
	"testing"

	"github.com/anchore/stereoscope"
	"github.com/anchore/stereoscope/pkg/image"
)

// Load reads the named fixture through stereoscope, with the source matching its format (a legacy save only loads
// as a docker-archive and a plain OCI layout only as an oci-archive). The image is cleaned up with the test.
func Load(t testing.TB, name string) *image.Image {
	t.Helper()
	return load(t, name, Path(t, name), Fixtures()[name].Format)
}

// LoadSpec is Load for a spec that is not a named fixture, for a test specific image that needs no goldens.
func LoadSpec(t testing.TB, spec Spec) *image.Image {
	t.Helper()
	return load(t, "spec", writeSpec(t, "spec", spec), spec.Format)
}

func load(t testing.TB, name, path string, format Format) *image.Image {
	t.Helper()
	source := image.DockerTarballSource
	if format == OCIArchive {
		source = image.OciTarballSource
	}
	img, err := stereoscope.GetImageFromSource(context.Background(), path, source)
	if err != nil {
		t.Fatalf("loading fixture %q: %v", name, err)
	}
	t.Cleanup(func() {
		if err := img.Cleanup(); err != nil {
			t.Errorf("cleaning up fixture %q: %v", name, err)
		}
	})
	return img
}

// Reference is the dive reference (scheme included) that loads the named fixture, for tests that go through dive's
// own source resolution.
func Reference(t testing.TB, name string) string {
	t.Helper()
	if spec, ok := Fixtures()[name]; ok && spec.Format == OCIArchive {
		return "oci-archive://" + Path(t, name)
	}
	return "docker-archive://" + Path(t, name)
}
