package podman

import (
	"bufio"
	"context"
	"fmt"
	"github.com/containers/libpod/libpod"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/dive/image/docker"
	"io/ioutil"
	"os"
)

type handler struct {
	id        string
	// note: podman supports saving docker formatted archives, we're leveraging this here
	// todo: add oci parser and image/layer objects
	image     docker.Image
}

func NewHandler() *handler {
	return &handler{}
}

func (handler *handler) Get(id string) error {
	handler.id = id

	path, err := handler.fetchArchive()
	if err != nil {
		return err
	}
	defer os.Remove(path)

	file, err := os.Open(path)

	// we use podman to extract a docker-formatted image
	img, err := docker.NewImageFromArchive(ioutil.NopCloser(bufio.NewReader(file)))
	if err != nil {
		return err
	}

	handler.image = img
	return nil
}

func (handler *handler) Build(args []string) (string, error) {
	var err error
	handler.id, err = buildImageFromCli(args)
	return handler.id, err
}

func (handler *handler) fetchArchive() (string, error) {
	var err error
	var ctx = context.Background()

	runtime, err := libpod.NewRuntime(ctx)
	if err != nil {
		return "", err
	}

	images, err := runtime.ImageRuntime().GetImages()
	if err != nil {
		return "", err
	}

	for _, item:= range images {
		for _, name := range item.Names() {
			if name == handler.id {
				file, err := ioutil.TempFile(os.TempDir(), "dive-handler-tar")
				if err != nil {
					return "", err
				}

				err = item.Save(ctx, "dive-export", "docker-archive", file.Name(), []string{}, false, false)
				if err != nil {
					return "", err
				}

				return file.Name(), nil
			}
		}
	}

	return "", fmt.Errorf("image could not be found")
}

func (handler *handler) Analyze() (*image.AnalysisResult, error) {
	return handler.image.Analyze()
}
