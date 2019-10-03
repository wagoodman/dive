package docker

import (
	"fmt"
	"github.com/wagoodman/dive/dive/image"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

var dockerVersion string

type handler struct {
	id     string
	client *client.Client
	image  Image
}

func NewHandler() *handler {
	return &handler{}
}

func (handler *handler) Get(id string) error {
	handler.id = id

	reader, err := handler.fetchArchive()
	if err != nil {
		return err
	}
	defer reader.Close()

	img, err := NewImageFromArchive(reader)
	if err != nil {
		return err
	}
	handler.image = img

	return nil
}

func (handler *handler) Build(args []string) (string, error) {
	var err error
	handler.id, err = buildImageFromCli(args)
	return handler.id, err
}

func (handler *handler) fetchArchive() (io.ReadCloser, error) {
	var err error

	// pull the handler if it does not exist
	ctx := context.Background()

	host := os.Getenv("DOCKER_HOST")
	var clientOpts []client.Opt

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
	handler.client, err = client.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, err
	}
	_, _, err = handler.client.ImageInspectWithRaw(ctx, handler.id)
	if err != nil {
		// don't use the API, the CLI has more informative output
		fmt.Println("Handler not available locally. Trying to pull '" + handler.id + "'...")
		err = runDockerCmd("pull", handler.id)
		if err != nil {
			return nil, err
		}
	}

	readCloser, err := handler.client.ImageSave(ctx, []string{handler.id})
	if err != nil {
		return nil, err
	}

	return readCloser, nil
}


func (handler *handler) Analyze() (*image.AnalysisResult, error) {
	return handler.image.Analyze()
}
