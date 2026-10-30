package source

import (
	"fmt"
	"os"
	"path/filepath"

	cliconfig "github.com/docker/cli/cli/config"
	dcontext "github.com/docker/cli/cli/context"
	ddocker "github.com/docker/cli/cli/context/docker"
	ctxstore "github.com/docker/cli/cli/context/store"
	"github.com/moby/moby/client"
)

// newDockerClient builds the docker API client handed to stereoscope, using dive's host resolution instead of
// stereoscope's (which only knows DOCKER_HOST and socket probing).
func newDockerClient() (client.APIClient, error) {
	configDir := cliconfig.Dir()
	ep, fromEnv, err := dockerEndpoint(configDir)
	if err != nil {
		return nil, fmt.Errorf("could not determine docker host: %w", err)
	}

	if fromEnv {
		if err := envTLS(&ep, configDir); err != nil {
			return nil, err
		}
	}
	// a context carries its own TLS material, ClientOpts applies it (and handles ssh hosts)
	opts, err := ep.ClientOpts()
	if err != nil {
		return nil, err
	}
	if fromEnv {
		// FromEnv goes first so the TLS client from ClientOpts wins over its own
		opts = append([]client.Opt{client.FromEnv}, opts...)
	}
	return client.New(opts...)
}

// envTLS gives a DOCKER_HOST or default host endpoint the TLS settings the docker cli derives from the environment:
// DOCKER_TLS_VERIFY turns on verified TLS, DOCKER_TLS alone unverified TLS (moby's FromEnv ignores DOCKER_TLS, so
// a client built from it alone talks plain http). The material comes from DOCKER_CERT_PATH, else the docker config
// dir. The client cert and key (and the CA when not verifying) are optional.
func envTLS(ep *ddocker.Endpoint, configDir string) error {
	verify := os.Getenv("DOCKER_TLS_VERIFY") != ""
	if !verify && os.Getenv("DOCKER_TLS") == "" {
		return nil
	}
	dir := os.Getenv("DOCKER_CERT_PATH")
	if dir == "" {
		dir = configDir
	}
	existing := func(name string) string {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err != nil {
			return ""
		}
		return p
	}
	ca := filepath.Join(dir, "ca.pem")
	if !verify {
		ca = existing("ca.pem")
	}
	data, err := dcontext.TLSDataFromFiles(ca, existing("cert.pem"), existing("key.pem"))
	if err != nil {
		return fmt.Errorf("could not load docker TLS material from %s: %w", dir, err)
	}
	ep.TLSData = data
	ep.SkipTLSVerify = !verify
	return nil
}

// dockerEndpoint determines the docker endpoint to connect to, in decreasing precedence:
//   - the DOCKER_HOST environment variable
//   - the context named by DOCKER_CONTEXT
//   - the currentContext in the docker config file
//   - the default docker host for the OS
//
// fromEnv is true for DOCKER_HOST and the default host, whose TLS settings come from the environment. A context
// endpoint comes with the context's TLS material.
func dockerEndpoint(configDir string) (ep ddocker.Endpoint, fromEnv bool, err error) {
	if h := os.Getenv("DOCKER_HOST"); h != "" {
		return ddocker.Endpoint{EndpointMeta: ddocker.EndpointMeta{Host: h}}, true, nil
	}
	defaultHost := ddocker.Endpoint{EndpointMeta: ddocker.EndpointMeta{Host: client.DefaultDockerHost}}

	name := os.Getenv("DOCKER_CONTEXT")
	if name == "" {
		cf, err := cliconfig.Load(configDir)
		if err != nil {
			return ddocker.Endpoint{}, false, err
		}
		name = cf.CurrentContext
	}

	// "default" is the docker cli's implicit context, it never lives in the store
	if name == "" || name == "default" {
		return defaultHost, true, nil
	}

	st := ctxstore.New(filepath.Join(configDir, "contexts"), ctxstore.NewConfig(
		func() any { return &ddocker.EndpointMeta{} },
		ctxstore.EndpointTypeGetter(ddocker.DockerEndpoint, func() any { return &ddocker.EndpointMeta{} }),
	))
	md, err := st.GetMetadata(name)
	if err != nil {
		return ddocker.Endpoint{}, false, err
	}

	// a context created with an empty host (`docker context create foo --docker "host="`) means the default host,
	// same as the docker cli
	meta, ok := md.Endpoints[ddocker.DockerEndpoint].(ddocker.EndpointMeta)
	if !ok || meta.Host == "" {
		return defaultHost, true, nil
	}
	ep, err = ddocker.WithTLSData(st, name, meta)
	return ep, false, err
}
