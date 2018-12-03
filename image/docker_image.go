package image

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/filetree"
	"github.com/wagoodman/dive/utils"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"strings"
)

var dockerVersion string

func newDockerImageAnalyzer() Analyzer {
	return &dockerImageAnalyzer{}
}

func newDockerImageManifest(manifestBytes []byte) dockerImageManifest {
	var manifest []dockerImageManifest
	err := json.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		logrus.Panic(err)
	}
	return manifest[0]
}

func newDockerImageConfig(configBytes []byte) dockerImageConfig {
	var imageConfig dockerImageConfig
	err := json.Unmarshal(configBytes, &imageConfig)
	if err != nil {
		logrus.Panic(err)
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

func (image *dockerImageAnalyzer) Parse(imageID string) error {
	var err error
	image.id = imageID
	// store discovered json files in a map so we can read the image in one pass
	image.jsonFiles = make(map[string][]byte)
	image.layerMap = make(map[string]*filetree.FileTree)

	// pull the image if it does not exist
	ctx := context.Background()
	image.client, err = client.NewClientWithOpts(client.WithVersion(dockerVersion), client.FromEnv)
	if err != nil {
		return err
	}
	_, _, err = image.client.ImageInspectWithRaw(ctx, imageID)
	if err != nil {
		// don't use the API, the CLI has more informative output
		fmt.Println("Image not available locally. Trying to pull '" + imageID + "'...")
		utils.RunDockerCmd("pull", imageID)
	}

	tarFile, _, err := image.getReader(imageID)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	err = image.read(tarFile)
	if err != nil {
		return err
	}
	return nil
}

// todo: it is bad that this is printing out to the screen
func (image *dockerImageAnalyzer) read(tarFile io.ReadCloser) error {
	tarReader := tar.NewReader(tarFile)

	var currentLayer uint
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			fmt.Println("    ╧")
			break
		}

		if err != nil {
			fmt.Println(err)
			utils.Exit(1)
		}

		layerProgress := fmt.Sprintf("[layer: %2d]", currentLayer)

		name := header.Name

		// some layer tars can be relative layer symlinks to other layer tars
		if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeReg {

			if strings.HasSuffix(name, "layer.tar") {
				currentLayer++
				if err != nil {
					return err
				}
				message := fmt.Sprintf("    ├─ %s %s ", layerProgress, "working...")
				fmt.Printf("\r%s", message)

				layerReader := tar.NewReader(tarReader)
				image.processLayerTar(name, layerReader, layerProgress)
			} else if strings.HasSuffix(name, ".json") {
				fileBuffer, err := ioutil.ReadAll(tarReader)
				if err != nil {
					return err
				}
				image.jsonFiles[name] = fileBuffer
			}
		}
	}

	return nil
}

func (image *dockerImageAnalyzer) Analyze() (*AnalysisResult, error) {
	image.trees = make([]*filetree.FileTree, 0)

	manifest := newDockerImageManifest(image.jsonFiles["manifest.json"])
	config := newDockerImageConfig(image.jsonFiles[manifest.ConfigPath])

	// build the content tree
	for _, treeName := range manifest.LayerTarPaths {
		image.trees = append(image.trees, image.layerMap[treeName])
	}

	// build the layers array
	image.layers = make([]*dockerLayer, len(image.trees))

	// note that the image config stores images in reverse chronological order, so iterate backwards through layers
	// as you iterate chronologically through history (ignoring history items that have no layer contents)
	layerIdx := len(image.trees) - 1
	tarPathIdx := 0
	for idx := 0; idx < len(config.History); idx++ {
		// ignore empty layers, we are only observing layers with content
		if config.History[idx].EmptyLayer {
			continue
		}

		tree := image.trees[(len(image.trees)-1)-layerIdx]
		config.History[idx].Size = uint64(tree.FileSize)

		image.layers[layerIdx] = &dockerLayer{
			history: config.History[idx],
			index:   layerIdx,
			tree:    image.trees[layerIdx],
			tarPath: manifest.LayerTarPaths[tarPathIdx],
		}

		layerIdx--
		tarPathIdx++
	}

	efficiency, inefficiencies := filetree.Efficiency(image.trees)

	layers := make([]Layer, len(image.layers))
	for i, v := range image.layers {
		layers[i] = v
	}

	return &AnalysisResult{
		Layers:         layers,
		RefTrees:       image.trees,
		Efficiency:     efficiency,
		Inefficiencies: inefficiencies,
	}, nil
}

func (image *dockerImageAnalyzer) getReader(imageID string) (io.ReadCloser, int64, error) {

	ctx := context.Background()
	result, _, err := image.client.ImageInspectWithRaw(ctx, imageID)
	if err != nil {
		return nil, -1, err
	}
	totalSize := result.Size

	readCloser, err := image.client.ImageSave(ctx, []string{imageID})
	if err != nil {
		return nil, -1, err
	}

	return readCloser, totalSize, nil
}

// todo: it is bad that this is printing out to the screen
func (image *dockerImageAnalyzer) processLayerTar(name string, reader *tar.Reader, layerProgress string) {
	tree := filetree.NewFileTree()
	tree.Name = name

	fileInfos := image.getFileList(reader)

	shortName := name[:15]
	pb := utils.NewProgressBar(int64(len(fileInfos)), 30)
	for idx, element := range fileInfos {
		tree.FileSize += uint64(element.TarHeader.FileInfo().Size())
		tree.AddPath(element.Path, element)

		if pb.Update(int64(idx)) {
			message := fmt.Sprintf("    ├─ %s %s : %s", layerProgress, shortName, pb.String())
			fmt.Printf("\r%s", message)
		}
	}
	pb.Done()
	message := fmt.Sprintf("    ├─ %s %s : %s", layerProgress, shortName, pb.String())
	fmt.Printf("\r%s\n", message)

	image.layerMap[tree.Name] = tree
}

// todo: it is bad that this is printing out to the screen
func (image *dockerImageAnalyzer) getFileList(tarReader *tar.Reader) []filetree.FileInfo {
	var files []filetree.FileInfo

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			utils.Exit(1)
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
