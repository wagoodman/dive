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

type resolver struct{}

func NewResolver() *resolver {
	return &resolver{}
}

func (r *resolver) Build(args []string) (*image.Image, error) {
	id, err := buildImageFromCli(args)
	if err != nil {
		return nil, err
	}
	return r.Fetch(id)
}

func (r *resolver) Fetch(id string) (*image.Image, error) {
	// todo: there are still a number of bugs remaining with this approach --stick with the docker archive for now
	// img, err := r.resolveFromDisk(id)
	// if err == nil {
	// 	return img, err
	// }

	img, err := r.resolveFromDockerArchive(id)
	if err == nil {
		return img, err
	}

	return nil, fmt.Errorf("unable to resolve image '%s'", id)
}

// func (r *resolver) resolveFromDisk(id string) (*image.Image, error) {
// 	var img *ImageDirectoryRef
// 	var err error
//
// 	runtime, err := libpod.NewRuntime(context.TODO())
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	images, err := runtime.ImageRuntime().GetImages()
// 	if err != nil {
// 		return nil, err
// 	}
//
// ImageLoop:
// 	for _, candidateImage := range images {
// 		for _, name := range candidateImage.Names() {
// 			if name == id {
// 				img, err = NewImageDirectoryRef(candidateImage)
// 				if err != nil {
// 					return nil, err
// 				}
// 				break ImageLoop
// 			}
// 		}
// 	}
//
// 	if img == nil {
// 		return nil, fmt.Errorf("could not find image by name: '%s'", id)
// 	}
//
// 	return img.ToImage()
// }

func (r *resolver) resolveFromDockerArchive(id string) (*image.Image, error) {
	path, err := r.fetchDockerArchive(id)
	if err != nil {
		return nil, err
	}
	defer os.Remove(path)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := docker.NewImageArchive(ioutil.NopCloser(bufio.NewReader(file)))
	if err != nil {
		return nil, err
	}
	return img.ToImage()
}

func (r *resolver) fetchDockerArchive(id string) (string, error) {
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

	for _, item := range images {
		for _, name := range item.Names() {
			if name == id {
				file, err := ioutil.TempFile(os.TempDir(), "dive-resolver-tar")
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
