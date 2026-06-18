//go:build !windows

package docker

import "testing"

func TestDefaultDockerHost(t *testing.T) {
	const want = "unix:///var/run/docker.sock"
	if defaultDockerHost != want {
		t.Errorf("defaultDockerHost = %q, want %q", defaultDockerHost, want)
	}
}
