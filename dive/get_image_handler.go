package dive

import (
	"fmt"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/dive/image/docker"
	"github.com/wagoodman/dive/dive/image/podman"
)

type Engine int

const (
	EngineUnknown Engine = iota
	EngineDocker
	EnginePodman
)

func (engine Engine) String() string {
	return [...]string{"unknown", "docker", "podman"}[engine]
}

var AllowedEngines = []string{EngineDocker.String(), EnginePodman.String()}

func GetEngine(engine string) Engine {
	switch engine {
	case "docker":
		return EngineDocker
	case "podman":
		return EnginePodman
	default:
		return EngineUnknown
	}
}

func GetImageHandler(engine Engine) (image.Handler, error) {
	switch engine {
	case EngineDocker:
		return docker.NewHandler(), nil
	case EnginePodman:
		return podman.NewHandler(), nil
	}

	return nil, fmt.Errorf("unable to determine image provider")
}
