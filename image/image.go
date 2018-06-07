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
	"github.com/wagoodman/docker-image-explorer/filetree"
	"golang.org/x/net/context"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type ImageManifest struct {
	Config   string
	RepoTags []string
	Layers   []string
}

func NewManifest(reader *tar.Reader, header *tar.Header) ImageManifest {
	size := header.Size
	manifestBytes := make([]byte, size)
	_, err := reader.Read(manifestBytes)
	if err != nil {
		panic(err)
	}
	var m []ImageManifest
	err = json.Unmarshal(manifestBytes, &m)
	if err != nil {
		panic(err)
	}
	return m[0]
}

func InitializeData(imageID string) (*ImageManifest, []*filetree.FileTree) {
	imageTarPath, tmpDir := saveImage(imageID)

	f, err := os.Open(imageTarPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer f.Close()
	defer os.RemoveAll(tmpDir)

	tarReader := tar.NewReader(f)
	targetName := "manifest.json"
	var manifest ImageManifest
	var layerMap map[string]*filetree.FileTree
	layerMap = make(map[string]*filetree.FileTree)

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
		if name == targetName {
			manifest = NewManifest(tarReader, header)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			if strings.HasSuffix(name, "layer.tar") {
				tree := filetree.NewFileTree()
				tree.Name = name
				fileInfos := getFileList(tarReader, header)
				for _, element := range fileInfos {
					tree.AddPath(element.Path, &element)
				}
				layerMap[tree.Name] = tree
			}
		default:
			fmt.Printf("ERRG: unknown tar entry: %v: %s\n", header.Typeflag, name)
		}
	}
	var trees []*filetree.FileTree
	trees = make([]*filetree.FileTree, 0)
	for _, treeName := range manifest.Layers {
		trees = append(trees, layerMap[treeName])
	}

	return &manifest, trees
}

func saveImage(imageID string) (string, string) {
	ctx := context.Background()
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	readCloser, err := dockerClient.ImageSave(ctx, []string{imageID})
	check(err)
	defer readCloser.Close()

	tmpDir, err := ioutil.TempDir("", "docker-image-explorer")
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

func getFileList(parentReader *tar.Reader, h *tar.Header) []filetree.FileChangeInfo {
	var files []filetree.FileChangeInfo
	size := h.Size
	tarredBytes := make([]byte, size)
	_, err := parentReader.Read(tarredBytes)
	if err != nil {
		panic(err)
	}
	r := bytes.NewReader(tarredBytes)
	tarReader := tar.NewReader(r)
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
			files = append(files, filetree.NewFileChangeInfo(tarReader, header, name))
		}
	}
	return files
}
