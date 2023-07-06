package docker

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"

	"github.com/wagoodman/dive/dive/image"
)

type engineResolver struct{}

func NewResolverFromEngine() *engineResolver {
	return &engineResolver{}
}

func (r *engineResolver) Fetch(id string) (*image.Image, error) {
	reader, err := r.fetchArchive(id)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	img, err := NewImageArchive(reader)
	if err != nil {
		return nil, err
	}
	return img.ToImage()
}

func (r *engineResolver) Build(args []string) (*image.Image, error) {
	id, err := buildImageFromCli(args)
	if err != nil {
		return nil, err
	}
	return r.Fetch(id)
}

func (r *engineResolver) fetchArchive(id string) (io.ReadCloser, error) {
	var err error
	var dockerClient *client.Client

	// pull the engineResolver if it does not exist
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

	clientOpts = append(clientOpts, client.WithAPIVersionNegotiation())
	dockerClient, err = client.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, err
	}
	_, _, err = dockerClient.ImageInspectWithRaw(ctx, id)
	if err != nil {
		// don't use the API, the CLI has more informative output
		fmt.Println("Handler not available locally. Trying to pull '" + id + "'...")
		err = runDockerCmd("pull", id)
		if err != nil {
			return nil, err
		}
	}

	readCloser, err := dockerClient.ImageSave(ctx, []string{id})
	if err != nil {
		return nil, err
	}

	return readCloser, nil
}
