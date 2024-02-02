//go:build !windows

package docker

const (
	defaultDockerHost = "unix:///var/run/docker.sock"
)
