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

type resolver struct {}

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
	img, err := r.resolveFromDisk(id)
	if err == nil {
		return img, err
	}
	img, err = r.resolveFromArchive(id)
	if err == nil {
		return img, err
	}

	return nil, fmt.Errorf("unable to resolve image '%s'", id)
}

func (r *resolver) resolveFromDisk(id string) (*image.Image, error) {
	// var err error
	return nil, fmt.Errorf("not implemented")
	//
	// runtime, err := libpod.NewRuntime(context.TODO())
	// if err != nil {
	// 	return nil, err
	// }
	//
	// images, err := runtime.ImageRuntime().GetImages()
	// if err != nil {
	// 	return nil, err
	// }
	//
	// // cfg, _ := runtime.GetConfig()
	// // cfg.StorageConfig.GraphRoot
	//
	// for _, item:= range images {
	// 	for _, name := range item.Names() {
	// 		if name == id {
	// 			fmt.Println("Found it!")
	//
	// 			curImg := item
	// 			for {
	// 				h, _ := curImg.History(context.TODO())
	// 				fmt.Printf("%+v %+v %+v\n", curImg.ID(), h[0].Size, h[0].CreatedBy)
	// 				x, _ := curImg.DriverData()
	// 				fmt.Printf("   %+v\n", x.Data["UpperDir"])
	//
	//
	// 				curImg, err = curImg.GetParent(context.TODO())
	// 				if err != nil || curImg == nil {
	// 					break
	// 				}
	// 			}
	//
	// 		}
	// 	}
	// }
	//
	// // os.Exit(0)
	// return nil, nil
}

func (r *resolver) resolveFromArchive(id string) (*image.Image, error) {
	path, err := r.fetchArchive(id)
	if err != nil {
		return nil, err
	}
	defer os.Remove(path)

	file, err := os.Open(path)
	defer file.Close()

	img, err := docker.NewImageFromArchive(ioutil.NopCloser(bufio.NewReader(file)))
	if err != nil {
		return nil, err
	}
	return img.ToImage()
}

func (r *resolver) fetchArchive(id string) (string, error) {
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
