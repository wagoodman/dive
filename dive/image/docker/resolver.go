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

type resolver struct {
	id     string
	client *client.Client
}

func NewResolver() *resolver {
	return &resolver{}
}

func (r *resolver) Resolve(id string) (image.Analyzer, error) {
	r.id = id

	reader, err := r.fetchArchive()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	img, err := NewImageFromArchive(reader)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func (r *resolver) Build(args []string) (string, error) {
	var err error
	r.id, err = buildImageFromCli(args)
	return r.id, err
}

func (r *resolver) fetchArchive() (io.ReadCloser, error) {
	var err error

	// pull the resolver if it does not exist
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
	r.client, err = client.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, err
	}
	_, _, err = r.client.ImageInspectWithRaw(ctx, r.id)
	if err != nil {
		// don't use the API, the CLI has more informative output
		fmt.Println("Handler not available locally. Trying to pull '" + r.id + "'...")
		err = runDockerCmd("pull", r.id)
		if err != nil {
			return nil, err
		}
	}

	readCloser, err := r.client.ImageSave(ctx, []string{r.id})
	if err != nil {
		return nil, err
	}

	return readCloser, nil
}

