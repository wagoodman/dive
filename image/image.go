package image

import (
	"archive/tar"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/client"
	"github.com/wagoodman/dive/filetree"
	"golang.org/x/net/context"
)

const (
	LayerFormat = "%-25s %5s %7s %s"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type ImageManifest struct {
	ConfigPath    string   `json:"Config"`
	RepoTags      []string `json:"RepoTags"`
	LayerTarPaths []string `json:"Layers"`
}

type ImageConfig struct {
	History []ImageHistoryEntry `json:"history"`
	RootFs RootFs `json:"rootfs"`
}

type RootFs struct {
	Type string `json:"type"`
	DiffIds []string `json:"diff_ids"`
}

type ImageHistoryEntry struct {
	ID string
	Size uint64
	Created string `json:"created"`
	Author string `json:"author"`
	CreatedBy string `json:"created_by"`
	EmptyLayer bool `json:"empty_layer"`
}

func NewImageManifest(reader *tar.Reader, header *tar.Header) ImageManifest {
	size := header.Size
	manifestBytes := make([]byte, size)
	_, err := reader.Read(manifestBytes)
	if err != nil && err != io.EOF {
		panic(err)
	}
	var manifest []ImageManifest
	err = json.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		panic(err)
	}
	return manifest[0]
}

func NewImageConfig(reader *tar.Reader, header *tar.Header) ImageConfig {
	size := header.Size
	configBytes := make([]byte, size)
	_, err := reader.Read(configBytes)
	if err != nil && err != io.EOF {
		panic(err)
	}
	var imageConfig ImageConfig
	err = json.Unmarshal(configBytes, &imageConfig)
	if err != nil {
		panic(err)
	}

	layerIdx := 0
	for idx := range imageConfig.History {
		if imageConfig.History[idx].EmptyLayer {
			imageConfig.History[idx].ID = "<missing>"
		} else {
			imageConfig.History[idx].ID = imageConfig.RootFs.DiffIds[layerIdx]
			layerIdx++
		}
	}

	return imageConfig
}

func GetImageConfig(imageTarPath string, manifest ImageManifest) ImageConfig{
	var config ImageConfig
	// read through the image contents and build a tree
	fmt.Println("Fetching image config...")
	tarFile, err := os.Open(imageTarPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer tarFile.Close()

	tarReader := tar.NewReader(tarFile)
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		name := header.Name
		if name == manifest.ConfigPath {
			config = NewImageConfig(tarReader, header)
		}
	}

	// obtain the image history
	return config
}

func InitializeData(imageID string) ([]*Layer, []*filetree.FileTree) {
	var manifest ImageManifest
	var layerMap = make(map[string]*filetree.FileTree)
	var trees []*filetree.FileTree = make([]*filetree.FileTree, 0)

	// save this image to disk temporarily to get the content info
	fmt.Println("Fetching image...")
	imageTarPath, tmpDir := saveImage(imageID)
	// imageTarPath := "/tmp/dive229500681/image.tar"
	// tmpDir := "/tmp/dive229500681"
	// fmt.Println(tmpDir)
	defer os.RemoveAll(tmpDir)

	// read through the image contents and build a tree
	fmt.Printf("Reading image '%s'...\n", imageID)
	tarFile, err := os.Open(imageTarPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer tarFile.Close()

	tarReader := tar.NewReader(tarFile)
	for {
		header, err := tarReader.Next()

		// log.Debug(header.Name)

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		name := header.Name
		if name == "manifest.json" {
			manifest = NewImageManifest(tarReader, header)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			// todo: process this loop in parallel, visualize with jotframe
			if strings.HasSuffix(name, "layer.tar") {
				tree := filetree.NewFileTree()
				tree.Name = name
				fileInfos := getFileList(tarReader, header)
				for _, element := range fileInfos {
					tree.FileSize += uint64(element.TarHeader.FileInfo().Size())
					tree.AddPath(element.Path, element)
				}
				layerMap[tree.Name] = tree
			}
		default:
			fmt.Printf("ERRG: unknown tar entry: %v: %s\n", header.Typeflag, name)
		}
	}

	// obtain the image history
	config := GetImageConfig(imageTarPath, manifest)

	// build the content tree
	fmt.Println("Building tree...")
	for _, treeName := range manifest.LayerTarPaths {
		trees = append(trees, layerMap[treeName])
	}


	// build the layers array
	layers := make([]*Layer, len(trees))

	// note that the image config stores images in reverse chronological order, so iterate backwards through layers
	// as you iterate chronologically through history (ignoring history items that have no layer contents)
	layerIdx := len(trees)-1
	for idx := 0; idx < len(config.History); idx++ {
		// ignore empty layers, we are only observing layers with content
		if config.History[idx].EmptyLayer {
			continue
		}

		config.History[idx].Size = uint64(trees[(len(trees)-1)-layerIdx].FileSize)

		layers[layerIdx] = &Layer{
			History: config.History[idx],
			Index: layerIdx,
			Tree: trees[layerIdx],
			RefTrees: trees,
		}

		if len(manifest.LayerTarPaths) > idx {
			layers[layerIdx].TarPath = manifest.LayerTarPaths[layerIdx]
		}
		layerIdx--
	}

	return layers, trees
}

func saveImage(imageID string) (string, string) {
	ctx := context.Background()
	dockerClient, err := client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}

	readCloser, err := dockerClient.ImageSave(ctx, []string{imageID})
	check(err)
	defer readCloser.Close()

	tmpDir, err := ioutil.TempDir("", "dive")
	check(err)

	imageTarPath := filepath.Join(tmpDir, "image.tar")
	imageFile, err := os.Create(imageTarPath)
	check(err)

	defer func() {
		if err := imageFile.Close(); err != nil {
			panic(err)
		}
	}()
	imageWriter := bufio.NewWriter(imageFile)

	buf := make([]byte, 1024)
	for {
		n, err := readCloser.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		if _, err := imageWriter.Write(buf[:n]); err != nil {
			panic(err)
		}
	}

	if err = imageWriter.Flush(); err != nil {
		panic(err)
	}

	return imageTarPath, tmpDir
}

func getFileList(parentReader *tar.Reader, header *tar.Header) []filetree.FileInfo {
	var files []filetree.FileInfo
	var tarredBytes = make([]byte, header.Size)

	_, err := parentReader.Read(tarredBytes)
	if err != nil && err != io.EOF {
		panic(err)
	}
	reader := bytes.NewReader(tarredBytes)
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeXGlobalHeader:
			fmt.Printf("ERRG: XGlobalHeader: %v: %s\n", header.Typeflag, name)
		case tar.TypeXHeader:
			fmt.Printf("ERRG: XHeader: %v: %s\n", header.Typeflag, name)
		default:
			files = append(files, filetree.NewFileInfo(tarReader, header, name))
		}
	}
	return files
}
