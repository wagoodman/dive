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

func main() {
	f, err := os.Open("image/cache.tar")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	// gzf, err := gzip.NewReader(f)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	tarReader := tar.NewReader(f)
	targetName := "manifest.json"
	var m Manifest
	var layerMap map[string]*FileTree
	layerMap = make(map[string]*FileTree)
	// var trees []*FileTree
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
			m = handleManifest(tarReader, header)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			//fmt.Println("File: ", name)
			if strings.HasSuffix(name, "layer.tar") {
				fmt.Println("Containing:")
				tree := NewTree()
				tree.name = name
				fmt.Printf("%s\n", tree.name)
				fileInfos := getFileList(tarReader, header)
				for _, element := range fileInfos {
					tree.AddPath(element.path, element)
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
	for _, treeName := range m.Layers {
		fmt.Printf("%s\n", treeName)
		trees = append(trees, layerMap[treeName])
	}

	for ix := range trees {
		if ix > 0 {
			trees[0].Stack(trees[ix])
		}
	}
	// fmt.Printf("%+v\n", m)
	// fmt.Printf("%+v\n", layerMap)
	fmt.Printf("%+v\n", trees)
	// fmt.Printf("Tree 1 is: \n %+v\n", trees[1])
	// fmt.Printf("Tree 2 is: \n %+v\n", trees[2])
	// trees[1].Stack(trees[2])
	// fmt.Printf("Tree 1 stacked with tree 2 is: \n %+v\n", trees[1])
	// fmt.Printf("The whle stack is \n %+v \n", trees[0])
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
			md5sum:   zeros,
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
	}
}

var zeros = [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

type FileChangeInfo struct {
	path     string
	typeflag byte
	md5sum   [16]byte
	diffType *DiffType
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
