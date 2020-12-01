package skopeo

import (
	"archive/tar"
	"compress/gzip"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image/docker"
	"io/ioutil"
	"os"
	"path/filepath"
)

func directoryToImageArchive(dir string) (*docker.ImageArchive, error) {
	img := &docker.ImageArchive{
		LayerMap: make(map[string]*filetree.FileTree),
	}

	bin, err := ioutil.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	img.Manifest = newManifest(bin)

	bin, err = ioutil.ReadFile(filepath.Join(dir, img.Manifest.ConfigPath))
	if err != nil {
		return nil, err
	}
	img.Config = docker.NewConfig(bin)

	for _, layer := range img.Manifest.LayerTarPaths {
		fd, err := os.Open(filepath.Join(dir, layer))
		if err != nil {
			return nil, err
		}
		defer fd.Close()

		gzipReader, err := gzip.NewReader(fd)
		if err != nil {
			return nil, err
		}
		tree, err := docker.ProcessLayerTar(layer, tar.NewReader(gzipReader))
		if err != nil {
			return nil, err
		}
		img.LayerMap[tree.Name] = tree
	}

	return img, nil
}
