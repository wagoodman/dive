package dive

import (
	"fmt"
	"github.com/wagoodman/dive/dive/v1/image"
	docker2 "github.com/wagoodman/dive/dive/v1/image/docker"
	"github.com/wagoodman/dive/dive/v1/image/podman"
	"strings"
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
	case SourceDockerEngine.String():
		return SourceDockerEngine
	case SourcePodmanEngine.String():
		return SourcePodmanEngine
	case SourceDockerArchive.String():
		return SourceDockerArchive
	case "docker-tar":
		return SourceDockerArchive
	default:
		return SourceUnknown
	}
}

func DeriveImageSource(image string) (ImageSource, string) {
	s := strings.SplitN(image, "://", 2)
	if len(s) < 2 {
		return SourceUnknown, ""
	}
	scheme, imageSource := s[0], s[1]

	switch scheme {
	case SourceDockerEngine.String():
		return SourceDockerEngine, imageSource
	case SourcePodmanEngine.String():
		return SourcePodmanEngine, imageSource
	case SourceDockerArchive.String():
		return SourceDockerArchive, imageSource
	case "docker-tar":
		return SourceDockerArchive, imageSource
	}
	return SourceUnknown, ""
}

func GetImageResolver(r ImageSource) (image.Resolver, error) {
	switch r {
	case SourceDockerEngine:
		return docker2.NewResolverFromEngine(), nil
	case SourcePodmanEngine:
		return podman.NewResolverFromEngine(), nil
	case SourceDockerArchive:
		return docker2.NewResolverFromArchive(), nil
	}

	return nil, fmt.Errorf("unable to determine image resolver")
}
