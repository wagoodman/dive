package source

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/moby/moby/client"
	"github.com/pelletier/go-toml/v2"
)

// podman's containers.conf, least to most relevant
var podmanConfigs = []string{
	"/usr/share/containers/containers.conf",
	"/etc/containers/containers.conf",
	filepath.Join(xdg.Home, ".config", "containers", "containers.conf"),
}

// newPodmanClient builds the docker API client for the podman provider. It finds the socket the way stereoscope's
// podman.GetClient does, minus its ssh fallback: that connection skips host key verification unless known_hosts has an
// exact unhashed entry for the host (never for a host with a port), and a pull over it sends registry credentials.
func newPodmanClient() (client.APIClient, error) {
	host, err := podmanHost(podmanConfigs, xdg.RuntimeDir, podmanMachineSocket)
	if err != nil {
		return nil, err
	}
	return client.New(client.WithHost(host))
}

// podmanHost resolves the podman API address: CONTAINER_HOST, then a unix destination in containers.conf, then the
// rootless and rootful default sockets, then the podman machine's API socket. ssh addresses are refused.
func podmanHost(configs []string, runtimeDir string, machineSocket func() string) (string, error) {
	host := os.Getenv("CONTAINER_HOST")
	if host == "" {
		for _, p := range configs {
			// later files win
			if h := podmanConfigSocket(p); h != "" {
				host = h
			}
		}
	}
	if host == "" {
		for _, p := range []string{filepath.Join(runtimeDir, "podman", "podman.sock"), "/run/podman/podman.sock"} {
			if _, err := os.Stat(p); err == nil {
				host = "unix://" + p
				break
			}
		}
	}
	if host == "" {
		// off linux podman runs in a machine, whose containers.conf entry is ssh only
		if p := machineSocket(); p != "" {
			if _, err := os.Stat(p); err == nil {
				host = "unix://" + p
			}
		}
	}
	if host == "" {
		return "", errors.New("no podman socket found (is the podman machine running? dive does not connect to podman over ssh), set CONTAINER_HOST=unix://<path to the podman socket>")
	}
	if u, err := url.Parse(host); err == nil && u.Scheme == "ssh" {
		return "", fmt.Errorf("refusing podman connection %q: dive does not connect to podman over ssh, since the host key would not be verified; set CONTAINER_HOST=unix://<path to a podman socket> (e.g. one forwarded with `ssh -L`)", host)
	}
	return host, nil
}

// podmanMachineSocket asks podman for the default machine's API socket, which is how podman is reached on macOS.
// Empty when podman or the machine isn't there.
func podmanMachineSocket() string {
	out, err := exec.Command("podman", "machine", "inspect", "--format", "{{.ConnectionInfo.PodmanSocket.Path}}").Output()
	if err != nil {
		return ""
	}
	path, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	return path
}

// podmanConfigSocket returns the unix destination of a containers.conf, the active service first.
func podmanConfigSocket(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var cc struct {
		Engine struct {
			ActiveService       string `toml:"active_service"`
			ServiceDestinations map[string]struct {
				URI string `toml:"uri"`
			} `toml:"service_destinations"`
		} `toml:"engine"`
	}
	if toml.Unmarshal(b, &cc) != nil {
		return ""
	}
	unix := func(uri string) string {
		if u, err := url.Parse(uri); err == nil && u.Scheme == "unix" {
			return "unix://" + u.Path
		}
		return ""
	}
	if h := unix(cc.Engine.ServiceDestinations[cc.Engine.ActiveService].URI); h != "" {
		return h
	}
	for _, d := range cc.Engine.ServiceDestinations {
		if h := unix(d.URI); h != "" {
			return h
		}
	}
	return ""
}
