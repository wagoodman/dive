package main

import (
	"archive/tar"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func initialize() {
	f, err := os.Open("image/cache.tar")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	tarReader := tar.NewReader(f)
	targetName := "manifest.json"
	var manifest Manifest
	var layerMap map[string]*FileTree
	layerMap = make(map[string]*FileTree)

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
				tree := NewTree()
				tree.name = name
				fmt.Printf("%s\n", tree.name)
				fileInfos := getFileList(tarReader, header)
				for _, element := range fileInfos {
					tree.AddPath(element.path, &element)
				}
				layerMap[tree.name] = tree
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
	var trees []*FileTree
	trees = make([]*FileTree, 0)
	for _, treeName := range manifest.Layers {
		trees = append(trees, layerMap[treeName])
	}

	data.manifest = &manifest
	data.refTrees = trees
	data.tree = StackRange(trees, 0)
}

func getFileList(parentReader *tar.Reader, h *tar.Header) []FileChangeInfo {
	var files []FileChangeInfo
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

func makeEntry(r *tar.Reader, h *tar.Header, path string) FileChangeInfo {
	if h.Typeflag == tar.TypeDir {
		return FileChangeInfo{
			path:     path,
			typeflag: h.Typeflag,
			md5sum:   [16]byte{},
		}
	}
	fileBytes := make([]byte, h.Size)
	_, err := r.Read(fileBytes)
	if err != nil && err != io.EOF {
		panic(err)
	}
	hash := md5.Sum(fileBytes)
	return FileChangeInfo{
		path:     path,
		typeflag: h.Typeflag,
		md5sum:   hash,
		diffType: Unchanged,
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
