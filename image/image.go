package image

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/docker/docker/client"
	"github.com/wagoodman/dive/filetree"
	"github.com/wagoodman/dive/utils"
	"golang.org/x/net/context"
)

// TODO: this file should be rethought... but since it's only for preprocessing it'll be tech debt for now.
var dockerVersion string

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type ProgressBar struct {
	percent    int
	rawTotal   int64
	rawCurrent int64
}

func NewProgressBar(total int64) *ProgressBar {
	return &ProgressBar{
		rawTotal: total,
	}
}

func (pb *ProgressBar) Done() {
	pb.rawCurrent = pb.rawTotal
	pb.percent = 100
}

func (pb *ProgressBar) Update(currentValue int64) (hasChanged bool) {
	pb.rawCurrent = currentValue
	percent := int(100.0 * (float64(pb.rawCurrent) / float64(pb.rawTotal)))
	if percent != pb.percent {
		hasChanged = true
	}
	pb.percent = percent
	return hasChanged
}

func (pb *ProgressBar) String() string {
	width := 40
	done := int((pb.percent * width) / 100.0)
	if done > width {
		done = width
	}
	todo := width - done
	if todo < 0 {
		todo = 0
	}
	head := 1

	return "[" + strings.Repeat("=", done) + strings.Repeat(">", head) + strings.Repeat(" ", todo) + "]" + fmt.Sprintf(" %d %% (%d/%d)", pb.percent, pb.rawCurrent, pb.rawTotal)
}

type ImageManifest struct {
	ConfigPath    string   `json:"Config"`
	RepoTags      []string `json:"RepoTags"`
	LayerTarPaths []string `json:"Layers"`
}

type ImageConfig struct {
	History []ImageHistoryEntry `json:"history"`
	RootFs  RootFs              `json:"rootfs"`
}

type RootFs struct {
	Type    string   `json:"type"`
	DiffIds []string `json:"diff_ids"`
}

type ImageHistoryEntry struct {
	ID         string
	Size       uint64
	Created    string `json:"created"`
	Author     string `json:"author"`
	CreatedBy  string `json:"created_by"`
	EmptyLayer bool   `json:"empty_layer"`
}

func NewImageManifest(manifestBytes []byte) ImageManifest {
	var manifest []ImageManifest
	err := json.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		logrus.Panic(err)
	}
	return manifest[0]
}

func NewImageConfig(configBytes []byte) ImageConfig {
	var imageConfig ImageConfig
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

func processLayerTar(layerMap map[string]*filetree.FileTree, name string, reader *tar.Reader, layerProgress string) {
	tree := filetree.NewFileTree()
	tree.Name = name

	fileInfos := getFileList(reader)

	shortName := name[:15]
	pb := NewProgressBar(int64(len(fileInfos)))
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

	layerMap[tree.Name] = tree
}

func InitializeData(imageID string) ([]*Layer, []*filetree.FileTree, float64, filetree.EfficiencySlice) {
	var layerMap = make(map[string]*filetree.FileTree)
	var trees = make([]*filetree.FileTree, 0)

	// pull the image if it does not exist
	ctx := context.Background()
	dockerClient, err := client.NewClientWithOpts(client.WithVersion(dockerVersion), client.FromEnv)
	if err != nil {
		fmt.Println("Could not connect to the Docker daemon:" + err.Error())
		utils.Exit(1)
	}
	_, _, err = dockerClient.ImageInspectWithRaw(ctx, imageID)
	if err != nil {
		// todo: check specific error
		// don't use the API, the CLI has more informative output
		fmt.Println("Image not available locally. Trying to pull '" + imageID + "'...")
		utils.RunDockerCmd("pull", imageID)
	}

	tarFile, _ := getImageReader(imageID)
	defer tarFile.Close()

	var currentLayer uint

	tarReader := tar.NewReader(tarFile)

	// json files are small. Let's store the in a map so we can read the image in one pass
	jsonFiles := make(map[string][]byte)

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
		var n int

		// some layer tars can be relative layer symlinks to other layer tars
		if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeReg {

			if strings.HasSuffix(name, "layer.tar") {
				currentLayer++
				if err != nil {
					logrus.Panic(err)
				}
				message := fmt.Sprintf("    ├─ %s %s ", layerProgress, "working...")
				fmt.Printf("\r%s", message)

				layerReader := tar.NewReader(tarReader)
				processLayerTar(layerMap, name, layerReader, layerProgress)
			} else if strings.HasSuffix(name, ".json") {
				var fileBuffer = make([]byte, header.Size)
				n, err = tarReader.Read(fileBuffer)
				if err != nil && err != io.EOF || int64(n) != header.Size {
					logrus.Panic(err)
				}
				jsonFiles[name] = fileBuffer
			}
		}
	}

	manifest := NewImageManifest(jsonFiles["manifest.json"])
	config := NewImageConfig(jsonFiles[manifest.ConfigPath])

	// build the content tree
	fmt.Println("  Building tree...")
	for _, treeName := range manifest.LayerTarPaths {
		trees = append(trees, layerMap[treeName])
	}

	// build the layers array
	layers := make([]*Layer, len(trees))

	// note that the image config stores images in reverse chronological order, so iterate backwards through layers
	// as you iterate chronologically through history (ignoring history items that have no layer contents)
	layerIdx := len(trees) - 1
	tarPathIdx := 0
	for idx := 0; idx < len(config.History); idx++ {
		// ignore empty layers, we are only observing layers with content
		if config.History[idx].EmptyLayer {
			continue
		}

		tree := trees[(len(trees)-1)-layerIdx]
		config.History[idx].Size = uint64(tree.FileSize)

		layers[layerIdx] = &Layer{
			History:  config.History[idx],
			Index:    layerIdx,
			Tree:     trees[layerIdx],
			RefTrees: trees,
			TarPath:  manifest.LayerTarPaths[tarPathIdx],
		}

		layerIdx--
		tarPathIdx++
	}

	fmt.Println("  Analyzing layers...")
	efficiency, inefficiencies := filetree.Efficiency(trees)

	return layers, trees, efficiency, inefficiencies
}

func getImageReader(imageID string) (io.ReadCloser, int64) {
	ctx := context.Background()
	dockerClient, err := client.NewClientWithOpts(client.WithVersion(dockerVersion), client.FromEnv)
	if err != nil {
		fmt.Println("Could not connect to the Docker daemon:" + err.Error())
		utils.Exit(1)
	}

	fmt.Println("  Fetching metadata...")

	result, _, err := dockerClient.ImageInspectWithRaw(ctx, imageID)
	check(err)
	totalSize := result.Size

	fmt.Println("  Fetching image...")

	readCloser, err := dockerClient.ImageSave(ctx, []string{imageID})
	check(err)

	return readCloser, totalSize
}

func getFileList(tarReader *tar.Reader) []filetree.FileInfo {
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
