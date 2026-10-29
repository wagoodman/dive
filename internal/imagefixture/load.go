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
	source := image.DockerTarballSource
	if spec, ok := Fixtures()[name]; ok && spec.Format == OCIArchive {
		source = image.OciTarballSource
	}
	img, err := stereoscope.GetImageFromSource(context.Background(), Path(t, name), source)
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
