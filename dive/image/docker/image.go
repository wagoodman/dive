package docker

import (
	"archive/tar"
	"fmt"
	"github.com/wagoodman/dive/dive/image"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/client"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/utils"
	"golang.org/x/net/context"
)

var dockerVersion string

type dockerImage struct {
	id        string
	client    *client.Client
	jsonFiles map[string][]byte
	trees     []*filetree.FileTree
	layerMap  map[string]*filetree.FileTree
	layers    []*dockerLayer
}

func NewDockerImage(imageId string) *dockerImage {
	return &dockerImage{
		// store discovered json files in a map so we can read the image in one pass
		jsonFiles: make(map[string][]byte),
		layerMap:  make(map[string]*filetree.FileTree),
		id:        imageId,
	}
}

func (img *dockerImage) Fetch() (io.ReadCloser, error) {
	var err error

	// pull the img if it does not exist
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

		if os.Getenv("DOCKER_TLS_VERIFY") != "" && os.Getenv("DOCKER_CERT_PATH") == "" {
			os.Setenv("DOCKER_CERT_PATH", "~/.docker")
		}

		clientOpts = append(clientOpts, client.FromEnv)
	}

	clientOpts = append(clientOpts, client.WithVersion(dockerVersion))
	img.client, err = client.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, err
	}
	_, _, err = img.client.ImageInspectWithRaw(ctx, img.id)
	if err != nil {
		// don't use the API, the CLI has more informative output
		fmt.Println("Image not available locally. Trying to pull '" + img.id + "'...")
		err = runDockerCmd("pull", img.id)
		if err != nil {
			return nil, err
		}
	}

	readCloser, err := img.client.ImageSave(ctx, []string{img.id})
	if err != nil {
		return nil, err
	}

	return readCloser, nil
}

func (img *dockerImage) Parse(tarFile io.ReadCloser) error {
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
				err := img.processLayerTar(name, currentLayer, layerReader)
				if err != nil {
					return err
				}
			} else if strings.HasSuffix(name, ".json") {
				fileBuffer, err := ioutil.ReadAll(tarReader)
				if err != nil {
					return err
				}
				img.jsonFiles[name] = fileBuffer
			}
		}
	}

	return nil
}

func (img *dockerImage) Analyze() (*image.AnalysisResult, error) {
	img.trees = make([]*filetree.FileTree, 0)

	manifest := newDockerImageManifest(img.jsonFiles["manifest.json"])
	config := newDockerImageConfig(img.jsonFiles[manifest.ConfigPath])

	// build the content tree
	for _, treeName := range manifest.LayerTarPaths {
		img.trees = append(img.trees, img.layerMap[treeName])
	}

	// build the layers array
	img.layers = make([]*dockerLayer, len(img.trees))

	// note that the img config stores images in reverse chronological order, so iterate backwards through layers
	// as you iterate chronologically through history (ignoring history items that have no layer contents)
	// Note: history is not required metadata in a docker img!
	tarPathIdx := 0
	histIdx := 0
	for layerIdx := len(img.trees) - 1; layerIdx >= 0; layerIdx-- {

		tree := img.trees[(len(img.trees)-1)-layerIdx]

		// ignore empty layers, we are only observing layers with content
		historyObj := imageHistoryEntry{
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

		img.layers[layerIdx] = &dockerLayer{
			history: historyObj,
			index:   tarPathIdx,
			tree:    img.trees[layerIdx],
			tarPath: manifest.LayerTarPaths[tarPathIdx],
		}
		img.layers[layerIdx].history.Size = tree.FileSize

		tarPathIdx++
	}

	efficiency, inefficiencies := filetree.Efficiency(img.trees)

	var sizeBytes, userSizeBytes uint64
	layers := make([]image.Layer, len(img.layers))
	for i, v := range img.layers {
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

	return &image.AnalysisResult{
		Layers:            layers,
		RefTrees:          img.trees,
		Efficiency:        efficiency,
		UserSizeByes:      userSizeBytes,
		SizeBytes:         sizeBytes,
		WastedBytes:       wastedBytes,
		WastedUserPercent: float64(wastedBytes) / float64(userSizeBytes),
		Inefficiencies:    inefficiencies,
	}, nil
}

func (img *dockerImage) processLayerTar(name string, layerIdx uint, reader *tar.Reader) error {
	tree := filetree.NewFileTree()
	tree.Name = name

	fileInfos, err := img.getFileList(reader)
	if err != nil {
		return err
	}

	for _, element := range fileInfos {
		tree.FileSize += uint64(element.Size)

		_, _, err := tree.AddPath(element.Path, element)
		if err != nil {
			return err
		}
	}

	img.layerMap[tree.Name] = tree
	return nil
}

func (img *dockerImage) getFileList(tarReader *tar.Reader) ([]filetree.FileInfo, error) {
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
