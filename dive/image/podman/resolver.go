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

type resolver struct {
	id        string
	// note: podman supports saving docker formatted archives, we're leveraging this here
	// todo: add oci parser and image/layer objects
	image     docker.Image
}

func NewResolver() *resolver {
	return &resolver{}
}

func (handler *resolver) Resolve(id string) (image.Analyzer, error) {
	handler.id = id

	path, err := handler.fetchArchive()
	if err != nil {
		return nil, err
	}
	defer os.Remove(path)

	file, err := os.Open(path)

	img, err := docker.NewImageFromArchive(ioutil.NopCloser(bufio.NewReader(file)))
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (handler *resolver) Build(args []string) (string, error) {
	var err error
	handler.id, err = buildImageFromCli(args)
	return handler.id, err
}

func (handler *resolver) fetchArchive() (string, error) {
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
				file, err := ioutil.TempFile(os.TempDir(), "dive-resolver-tar")
				if err != nil {
					return "", err
				}

				err = item.Save(ctx, "dive-export", "docker-archive", file.Name(), []string{}, false, false)
				if err != nil {
					return "", err
				}

				fmt.Println(file.Name())

				return file.Name(), nil
			}
		}
	}

	return "", fmt.Errorf("image could not be found")
}
