package skopeo

import (
	"fmt"
	"github.com/wagoodman/dive/dive/image"
	"io/ioutil"
	"os"
	"os/exec"
)

type resolver struct{}

func NewResolverFromEngine() *resolver {
	return &resolver{}
}

func (r *resolver) Build(args []string) (*image.Image, error) {
	return nil, fmt.Errorf("unsupported platform")
}

func (r *resolver) Fetch(path string) (*image.Image, error) {
	dir, err := ioutil.TempDir("", "skopeo")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	cmd := exec.Command("skopeo", "copy", "--override-os=linux", "docker://"+path, "dir:"+dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	img, err := directoryToImageArchive(dir)
	if err != nil {
		return nil, err
	}
	return img.ToImage()
}
