package source

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/connhelper"
	ddocker "github.com/docker/cli/cli/context/docker"
	ctxstore "github.com/docker/cli/cli/context/store"
	"github.com/moby/moby/client"
)

// newDockerClient builds the docker API client handed to stereoscope, using dive's host resolution instead of
// stereoscope's (which only knows DOCKER_HOST and socket probing).
func newDockerClient() (client.APIClient, error) {
	host, err := dockerHost(cliconfig.Dir())
	if err != nil {
		return nil, fmt.Errorf("could not determine docker host: %w", err)
	}

	opts := []client.Opt{client.FromEnv, client.WithHost(host)}
	if strings.HasPrefix(host, "ssh://") {
		helper, err := connhelper.GetConnectionHelper(host)
		if err != nil {
			return nil, fmt.Errorf("failed to get docker connection helper: %w", err)
		}
		opts = append(opts,
			client.WithHTTPClient(&http.Client{Transport: &http.Transport{DialContext: helper.Dialer}}),
			client.WithHost(helper.Host),
			client.WithDialContext(helper.Dialer),
		)
	} else if os.Getenv("DOCKER_TLS_VERIFY") != "" && os.Getenv("DOCKER_CERT_PATH") == "" {
		// same default as the docker cli: certs live in the docker config dir
		if err := os.Setenv("DOCKER_CERT_PATH", cliconfig.Dir()); err != nil {
			return nil, err
		}
	}

	return client.New(opts...)
}

// dockerHost determines the docker host to connect to, in decreasing precedence:
//   - the DOCKER_HOST environment variable
//   - the host of the context named by DOCKER_CONTEXT
//   - the host of the currentContext in the docker config file
//   - the default docker host for the OS
func dockerHost(configDir string) (string, error) {
	if h := os.Getenv("DOCKER_HOST"); h != "" {
		return h, nil
	}

	name := os.Getenv("DOCKER_CONTEXT")
	if name == "" {
		cf, err := cliconfig.Load(configDir)
		if err != nil {
			return "", err
		}
		name = cf.CurrentContext
	}

	// "default" is the docker cli's implicit context, it never lives in the store
	if name == "" || name == "default" {
		return client.DefaultDockerHost, nil
	}

	storeConfig := ctxstore.NewConfig(
		func() any { return &ddocker.EndpointMeta{} },
		ctxstore.EndpointTypeGetter(ddocker.DockerEndpoint, func() any { return &ddocker.EndpointMeta{} }),
	)
	md, err := ctxstore.New(filepath.Join(configDir, "contexts"), storeConfig).GetMetadata(name)
	if err != nil {
		return "", err
	}

	// a context created with an empty host (`docker context create foo --docker "host="`) means the default host,
	// same as the docker cli
	ep, ok := md.Endpoints[ddocker.DockerEndpoint].(ddocker.EndpointMeta)
	if !ok || ep.Host == "" {
		return client.DefaultDockerHost, nil
	}
	return ep.Host, nil
}
