//go:build windows

package docker

import "testing"

func TestDefaultDockerHost(t *testing.T) {
	// The default Windows pipe URL must match Moby's client.DefaultDockerHost
	// (https://github.com/moby/moby/blob/master/client/client_windows.go)
	// so that the docker client does not URL-encode the pipe path.
	const want = "npipe:////./pipe/docker_engine"
	if defaultDockerHost != want {
		t.Errorf("defaultDockerHost = %q, want %q", defaultDockerHost, want)
	}
}
