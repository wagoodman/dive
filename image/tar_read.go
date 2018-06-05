package image

import (
	"archive/tar"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/wagoodman/docker-image-explorer/filetree"
)

func InitializeData() (*Manifest, []*filetree.FileTree) {
	f, err := os.Open("./.image/cache.tar")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	tarReader := tar.NewReader(f)
	targetName := "manifest.json"
	var manifest Manifest
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
			manifest = handleManifest(tarReader, header)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:

			if strings.HasSuffix(name, "layer.tar") {
				fmt.Println("Containing:")
				tree := filetree.NewFileTree()
				tree.Name = name
				fmt.Printf("%s\n", tree.Name)
				fileInfos := getFileList(tarReader, header)
				for _, element := range fileInfos {
					tree.AddPath(element.Path, &element)
				}
				layerMap[tree.Name] = tree
			}
		default:
			fmt.Printf("%s : %c %s %s\n",
				"hmmm?",
				header.Typeflag,
				"in file",
				name,
			)
		}
	}
	var trees []*filetree.FileTree
	trees = make([]*filetree.FileTree, 0)
	for _, treeName := range manifest.Layers {
		trees = append(trees, layerMap[treeName])
	}

	return &manifest, trees
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
		case tar.TypeDir:
			files = append(files, makeEntry(tarReader, header, name))
		case tar.TypeReg:
			files = append(files, makeEntry(tarReader, header, name))
			continue
		case tar.TypeSymlink:
			files = append(files, makeEntry(tarReader, header, name))
		default:
			fmt.Printf("%s : %c %s %s\n",
				"hmmm?",
				header.Typeflag,
				"in file",
				name,
			)
		}
	}
	return files
}

func makeEntry(r *tar.Reader, h *tar.Header, path string) filetree.FileChangeInfo {
	if h.Typeflag == tar.TypeDir {
		return filetree.FileChangeInfo{
			Path:     path,
			Typeflag: h.Typeflag,
			MD5sum:   [16]byte{},
		}
	}
	fileBytes := make([]byte, h.Size)
	_, err := r.Read(fileBytes)
	if err != nil && err != io.EOF {
		panic(err)
	}
	hash := md5.Sum(fileBytes)
	return filetree.FileChangeInfo{
		Path:     path,
		Typeflag: h.Typeflag,
		MD5sum:   hash,
		DiffType: filetree.Unchanged,
	}
}

type Manifest struct {
	Config   string
	RepoTags []string
	Layers   []string
}

func handleManifest(r *tar.Reader, header *tar.Header) Manifest {
	size := header.Size
	manifestBytes := make([]byte, size)
	_, err := r.Read(manifestBytes)
	if err != nil {
		panic(err)
	}
	var m [1]Manifest
	err = json.Unmarshal(manifestBytes, &m)
	if err != nil {
		panic(err)
	}
	return m[0]
}
