package podman

import (
	"context"
	"fmt"
	podmanImage "github.com/containers/libpod/libpod/image"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"os"
	"path/filepath"
	"strings"
)

type ImageDirectoryRef struct {
	layerOrder []string
	treeMap    map[string]*filetree.FileTree
	layerMap   map[string]*podmanImage.Image
}

func NewImageDirectoryRef(img *podmanImage.Image) (*ImageDirectoryRef, error) {
	imgDirRef := &ImageDirectoryRef{
		layerOrder: make([]string, 0),
		treeMap:  make(map[string]*filetree.FileTree),
		layerMap: make(map[string]*podmanImage.Image),
	}

	ctx := context.TODO()

	curImg := img
	for {
		// h, _ := img.History(ctx)
		// fmt.Printf("%+v %+v %+v\n", img.ID(), h[0].Size, h[0].CreatedBy)

		driver, err := curImg.DriverData()
		if err != nil {
			return nil, fmt.Errorf("graph driver error: %+v", err)
		}

		if driver.Name != "overlay" {
			return nil, fmt.Errorf("unsupported graph driver: %s", driver.Name)
		}

		rootDir, exists := driver.Data["UpperDir"]
		if !exists {
			return nil, fmt.Errorf("graph has no upper dir")
		}
		
		if _, err := os.Stat(rootDir); os.IsNotExist(err) {
			return nil, fmt.Errorf("graph root dir does not exist: %s", rootDir)
		}

		// build tree from directory...
		tree, err := processLayer(curImg.ID(), rootDir)
		if err != nil {
			return nil, err
		}

		// record the tree and layer info
		imgDirRef.treeMap[curImg.ID()] = tree
		imgDirRef.layerMap[curImg.ID()] = curImg
		imgDirRef.layerOrder = append(imgDirRef.layerOrder, curImg.ID())

		// continue to the next image
		curImg, err = curImg.GetParent(ctx)
		if err != nil || curImg == nil {
			break
		}
	}

	return imgDirRef, nil
}

func processLayer(name, rootDir string) (*filetree.FileTree, error)  {
	tree := filetree.NewFileTree()
	tree.Name = name

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// add this file to the tree...
		fileInfo := filetree.NewFileInfo(path, "/"+strings.TrimPrefix(path, rootDir), info)

		tree.FileSize += uint64(fileInfo.Size)

		_, _, err = tree.AddPath(fileInfo.Path, fileInfo)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to walk upper directory tree")
	}

	return tree, nil
}

func (img *ImageDirectoryRef) ToImage() (*image.Image, error) {
	trees := make([]*filetree.FileTree, 0)
	// build the content tree
	// todo: this isn't needed!
	for _, id := range img.layerOrder {
		tr, exists := img.treeMap[id]
		if exists {
			trees = append(trees, tr)
			continue
		}
		return nil, fmt.Errorf("could not find '%s' in parsed trees", id)
	}

	layers := make([]image.Layer, len(trees))

	// note that the resolver config stores images in reverse chronological order, so iterate backwards through layers
	// as you iterate chronologically through history (ignoring history items that have no layer contents)
	// Note: history is not required metadata in a docker image!
	tarPathIdx := 0
	for layerIdx := len(trees) - 1; layerIdx >= 0; layerIdx-- {
		id := img.layerOrder[layerIdx]
		layers[layerIdx] = &layer{
			obj:     img.layerMap[id],
			index:   tarPathIdx,
			tree:    trees[layerIdx],
		}
	}

	return &image.Image{
		Trees: trees,
		Layers: layers,
	}, nil
}
