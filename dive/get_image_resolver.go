package dive

import (
	"fmt"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/dive/image/docker"
	"github.com/wagoodman/dive/dive/image/podman"
)

const (
	SourceUnknown ImageSource = iota
	SourceDockerEngine
	SourcePodmanEngine
	SourceDockerArchive
)

type ImageSource int

var ImageSources = []string{SourceDockerEngine.String(), SourcePodmanEngine.String(), SourceDockerArchive.String()}

func (r ImageSource) String() string {
	return [...]string{"unknown", "docker", "podman", "docker-archive"}[r]
}

func ParseImageSource(r string) ImageSource {
	switch r {
	case "docker":
		return SourceDockerEngine
	case "podman":
		return SourcePodmanEngine
	case "docker-archive":
		return SourceDockerArchive
	case "docker-tar":
		return SourceDockerArchive
	default:
		return SourceUnknown
	}
}

func GetImageResolver(r ImageSource) (image.Resolver, error) {
	switch r {
	case SourceDockerEngine:
		return docker.NewResolverFromEngine(), nil
	case SourcePodmanEngine:
		return podman.NewResolverFromEngine(), nil
	case SourceDockerArchive:
		return docker.NewResolverFromArchive(), nil
	}

	return nil, fmt.Errorf("unable to determine image resolver")
}
