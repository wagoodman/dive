package image

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/filetree"
	"github.com/wagoodman/dive/utils"
	"golang.org/x/net/context"
)

var dockerVersion string

func newDockerImageAnalyzer(imageId string) Analyzer {
	return &dockerImageAnalyzer{
		// store discovered json files in a map so we can read the image in one pass
		jsonFiles: make(map[string][]byte),
		layerMap:  make(map[string]*filetree.FileTree),
		id:        imageId,
	}
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

func (image *dockerImageAnalyzer) Fetch() (io.ReadCloser, error) {
	var err error

	// pull the image if it does not exist
	ctx := context.Background()

	host := os.Getenv("DOCKER_HOST")
	var clientOpts []func(*client.Client) error

	switch strings.Split(host, ":")[0] {
	case "ssh":
		helper, err := connhelper.GetConnectionHelper(host)
		if err != nil {
			fmt.Println("docker host", err)
		}
		clientOpts = append(clientOpts, func(c *client.Client) error {
			httpClient := &http.Client{
				Transport: &http.Transport{
					DialContext: helper.Dialer,
				},
			}
			return client.WithHTTPClient(httpClient)(c)
		})
		clientOpts = append(clientOpts, client.WithHost(helper.Host))
		clientOpts = append(clientOpts, client.WithDialContext(helper.Dialer))

	default:
		clientOpts = append(clientOpts, client.FromEnv)
	}

	clientOpts = append(clientOpts, client.WithVersion(dockerVersion))
	image.client, err = client.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, err
	}
	_, _, err = image.client.ImageInspectWithRaw(ctx, image.id)
	if err != nil {
		// don't use the API, the CLI has more informative output
		fmt.Println("Image not available locally. Trying to pull '" + image.id + "'...")
		utils.RunDockerCmd("pull", image.id)
	}

	readCloser, err := image.client.ImageSave(ctx, []string{image.id})
	if err != nil {
		return nil, err
	}

	return readCloser, nil
}

func (image *dockerImageAnalyzer) Parse(tarFile io.ReadCloser) error {
	tarReader := tar.NewReader(tarFile)

	var currentLayer uint
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

		// some layer tars can be relative layer symlinks to other layer tars
		if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeReg {

			if strings.HasSuffix(name, "layer.tar") {
				currentLayer++
				if err != nil {
					return err
				}
				layerReader := tar.NewReader(tarReader)
				err := image.processLayerTar(name, currentLayer, layerReader)
				if err != nil {
					return err
				}
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
	// Note: history is not required metadata in a docker image!
	tarPathIdx := 0
	histIdx := 0
	for layerIdx := len(image.trees) - 1; layerIdx >= 0; layerIdx-- {

		tree := image.trees[(len(image.trees)-1)-layerIdx]

		// ignore empty layers, we are only observing layers with content
		historyObj := dockerImageHistoryEntry{
			CreatedBy: "(missing)",
		}
		for nextHistIdx := histIdx; nextHistIdx < len(config.History); nextHistIdx++ {
			if !config.History[nextHistIdx].EmptyLayer {
				histIdx = nextHistIdx
				break
			}
		}
		if histIdx < len(config.History) && !config.History[histIdx].EmptyLayer {
			historyObj = config.History[histIdx]
			histIdx++
		}

		image.layers[layerIdx] = &dockerLayer{
			history: historyObj,
			index:   tarPathIdx,
			tree:    image.trees[layerIdx],
			tarPath: manifest.LayerTarPaths[tarPathIdx],
		}
		image.layers[layerIdx].history.Size = uint64(tree.FileSize)

		tarPathIdx++
	}

	efficiency, inefficiencies := filetree.Efficiency(image.trees)

	var sizeBytes, userSizeBytes uint64
	layers := make([]Layer, len(image.layers))
	for i, v := range image.layers {
		layers[i] = v
		sizeBytes += v.Size()
		if i != 0 {
			userSizeBytes += v.Size()
		}
	}

	var wastedBytes uint64
	for idx := 0; idx < len(inefficiencies); idx++ {
		fileData := inefficiencies[len(inefficiencies)-1-idx]
		wastedBytes += uint64(fileData.CumulativeSize)
	}

	return &AnalysisResult{
		Layers:            layers,
		RefTrees:          image.trees,
		Efficiency:        efficiency,
		UserSizeByes:      userSizeBytes,
		SizeBytes:         sizeBytes,
		WastedBytes:       wastedBytes,
		WastedUserPercent: float64(float64(wastedBytes) / float64(userSizeBytes)),
		Inefficiencies:    inefficiencies,
	}, nil
}

func (image *dockerImageAnalyzer) processLayerTar(name string, layerIdx uint, reader *tar.Reader) error {
	tree := filetree.NewFileTree()
	tree.Name = name

	fileInfos, err := image.getFileList(reader)
	if err != nil {
		return err
	}

	for _, element := range fileInfos {
		tree.FileSize += uint64(element.Size)

		tree.AddPath(element.Path, element)
	}

	image.layerMap[tree.Name] = tree
	return nil
}

func (image *dockerImageAnalyzer) getFileList(tarReader *tar.Reader) ([]filetree.FileInfo, error) {
	var files []filetree.FileInfo

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(err)
			utils.Exit(1)
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeXGlobalHeader:
			return nil, fmt.Errorf("unexptected tar file: (XGlobalHeader): type=%v name=%s", header.Typeflag, name)
		case tar.TypeXHeader:
			return nil, fmt.Errorf("unexptected tar file (XHeader): type=%v name=%s", header.Typeflag, name)
		default:
			files = append(files, filetree.NewFileInfo(tarReader, header, name))
		}
	}
	return files, nil
}
