package docker

import (
	"fmt"
	"github.com/spf13/afero"
	"io"
	"net/http"
	"os"
	"strings"

	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/connhelper"
	ddocker "github.com/docker/cli/cli/context/docker"
	ctxstore "github.com/docker/cli/cli/context/store"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"

	"github.com/wagoodman/dive/dive/image"
)

type engineResolver struct{}

func NewResolverFromEngine() *engineResolver {
	return &engineResolver{}
}

// Name returns the name of the resolver to display to the user.
func (r *engineResolver) Name() string {
	return "docker-engine"
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
	return img.ToImage(id)
}

func (r *engineResolver) Build(args []string) (*image.Image, error) {
	id, err := buildImageFromCli(afero.NewOsFs(), args)
	if err != nil {
		return nil, err
	}
	return r.Fetch(id)
}

func (r *engineResolver) Extract(id string, l string, p string) error {
	reader, err := r.fetchArchive(id)
	if err != nil {
		return err
	}

	if err := ExtractFromImage(io.NopCloser(reader), l, p); err == nil {
		return nil
	}

	return fmt.Errorf("unable to extract from image '%s': %+v", id, err)
}

func (r *engineResolver) fetchArchive(id string) (io.ReadCloser, error) {
	var err error
	var dockerClient *client.Client

	// pull the engineResolver if it does not exist
	ctx := context.Background()

	host, err := determineDockerHost()
	if err != nil {
		fmt.Printf("> could not determine docker host: %v\n", err)
	}
	clientOpts := []client.Opt{client.FromEnv}
	clientOpts = append(clientOpts, client.WithHost(host))

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

		clientOpts = append(clientOpts, client.WithHost(host))
		clientOpts = append(clientOpts, client.WithDialContext(helper.Dialer))

	default:

		if os.Getenv("DOCKER_TLS_VERIFY") != "" && os.Getenv("DOCKER_CERT_PATH") == "" {
			os.Setenv("DOCKER_CERT_PATH", "~/.docker")
		}
	}

	clientOpts = append(clientOpts, client.WithAPIVersionNegotiation())
	dockerClient, err = client.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, err
	}
	_, err = dockerClient.ImageInspect(ctx, id)
	if err != nil {
		// check if the error is due to the image not existing locally
		if client.IsErrNotFound(err) {
			fmt.Println("The image is not available locally. Trying to pull '" + id + "'...")
			err = runDockerCmd("pull", id)
			if err != nil {
				return nil, err
			}
		} else {
			// Some other error occurred, return it
			return nil, err
		}
	}

	readCloser, err := dockerClient.ImageSave(ctx, []string{id})
	if err != nil {
		return nil, err
	}

	return readCloser, nil
}

// determineDockerHost tries to the determine the docker host that we should connect to
// in the following order of decreasing precedence:
//   - value of "DOCKER_HOST" environment variable
//   - host retrieved from the current context (specified via DOCKER_CONTEXT)
//   - "default docker host" for the host operating system, otherwise
func determineDockerHost() (string, error) {
	// If the docker host is explicitly set via the "DOCKER_HOST" environment variable,
	// then its a no-brainer :shrug:
	if os.Getenv("DOCKER_HOST") != "" {
		return os.Getenv("DOCKER_HOST"), nil
	}

	currentContext := os.Getenv("DOCKER_CONTEXT")
	if currentContext == "" {
		cf, err := cliconfig.Load(cliconfig.Dir())
		if err != nil {
			return "", err
		}
		currentContext = cf.CurrentContext
	}

	if currentContext == "" {
		// If a docker context is neither specified via the "DOCKER_CONTEXT" environment variable nor via the
		// $HOME/.docker/config file, then we fall back to connecting to the "default docker host" meant for
		// the host operating system.
		return defaultDockerHost, nil
	}

	storeConfig := ctxstore.NewConfig(
		func() interface{} { return &ddocker.EndpointMeta{} },
		ctxstore.EndpointTypeGetter(ddocker.DockerEndpoint, func() interface{} { return &ddocker.EndpointMeta{} }),
	)

	st := ctxstore.New(cliconfig.ContextStoreDir(), storeConfig)
	md, err := st.GetMetadata(currentContext)
	if err != nil {
		return "", err
	}
	dockerEP, ok := md.Endpoints[ddocker.DockerEndpoint]
	if !ok {
		return "", err
	}
	dockerEPMeta, ok := dockerEP.(ddocker.EndpointMeta)
	if !ok {
		return "", fmt.Errorf("expected docker.EndpointMeta, got %T", dockerEP)
	}

	if dockerEPMeta.Host != "" {
		return dockerEPMeta.Host, nil
	}

	// We might end up here, if the context was created with the `host` set to an empty value (i.e. '').
	// For example:
	// ```sh
	// docker context create foo --docker "host="
	// ```
	// In such scenario, we mimic the `docker` cli and try to connect to the "default docker host".
	return defaultDockerHost, nil
}
