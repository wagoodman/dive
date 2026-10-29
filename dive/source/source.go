// Package source resolves a user provided image reference to a stereoscope image.
package source

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/anchore/stereoscope/pkg/image/containerd"
	"github.com/anchore/stereoscope/pkg/image/containerstorage"
	"github.com/anchore/stereoscope/pkg/image/docker"
	"github.com/anchore/stereoscope/pkg/image/oci"
	"github.com/anchore/stereoscope/pkg/image/podman"
)

// image sources selectable with a reference scheme or the --source flag
const (
	Docker            = "docker"
	Podman            = "podman"
	DockerArchive     = "docker-archive"
	OCIArchive        = "oci-archive"
	OCIDir            = "oci-dir"
	Registry          = "registry"
	Containerd        = "containerd"
	ContainersStorage = "containers-storage"
	// Auto lets stereoscope detect the source (files, then daemons and stores, then a registry pull). It is only
	// valid as a --source value, never as a reference scheme.
	Auto = "auto"
)

// Sources lists every valid --source value (aliases excluded).
var Sources = []string{Docker, Podman, DockerArchive, OCIArchive, OCIDir, Registry, Containerd, ContainersStorage, Auto}

var aliases = map[string]string{
	"docker-tar":   DockerArchive,
	"oci-registry": Registry,
}

// a "registry:5000/img" style reference is a host with a port, not a scheme
var portPattern = regexp.MustCompile(`^\d+(/|$)`)

// Config controls how Load resolves a reference.
type Config struct {
	// Source is used when the reference has no scheme. Empty means Docker.
	Source string
	// Platform selects an image from a multi-platform index (e.g. "linux/arm64"). Empty means the default.
	Platform string
}

// Normalize maps aliases to their canonical source name and reports whether the result is a valid source.
func Normalize(s string) (string, bool) {
	if a, ok := aliases[s]; ok {
		s = a
	}
	return s, slices.Contains(Sources, s)
}

// Parse splits a user reference into a source and the reference to hand to that source. Both "docker://img" and
// stereoscope's "docker:img" forms are accepted. Without a known scheme the (normalized) defaultSource is returned
// with the input untouched; an empty defaultSource means Docker.
func Parse(input, defaultSource string) (src, ref string) {
	if scheme, rest, ok := strings.Cut(input, "://"); ok {
		if s, ok := schemeSource(scheme); ok {
			return s, rest
		}
	}
	if scheme, rest, ok := strings.Cut(input, ":"); ok && !portPattern.MatchString(rest) {
		if s, ok := schemeSource(scheme); ok {
			return s, rest
		}
	}
	if defaultSource == "" {
		defaultSource = Docker
	}
	src, _ = Normalize(defaultSource)
	return src, input
}

func schemeSource(scheme string) (string, bool) {
	s, ok := Normalize(scheme)
	return s, ok && s != Auto
}

// Load resolves the user reference (see Parse) and reads the image with stereoscope. The caller owns the returned
// image and must call img.Cleanup() when done with it, including on interrupt, to remove the temp dirs it holds.
func Load(ctx context.Context, input string, cfg Config) (*image.Image, error) {
	src, ref := Parse(input, cfg.Source)

	var platform *image.Platform
	if cfg.Platform != "" {
		p, err := image.NewPlatform(cfg.Platform)
		if err != nil {
			return nil, err
		}
		platform = p
	}

	tmp := file.NewTempDirGenerator("dive")
	ps, err := providers(src, ref, runtime.GOOS, platform, tmp)
	if err != nil {
		return nil, err
	}

	var errs []error
	for _, p := range ps {
		img, err := p.Provide(ctx)
		if err == nil {
			return img, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", p.Name(), err))
	}
	// providers that failed may have left temp dirs behind
	errs = append(errs, tmp.Cleanup())
	return nil, fmt.Errorf("unable to load image %q from %s: %w", ref, src, errors.Join(errs...))
}

// providers returns the stereoscope providers to try, in order, for the given source. Constructing them is cheap,
// nothing connects until Provide.
func providers(src, ref, goos string, platform *image.Platform, tmp *file.TempDirGenerator) ([]image.Provider, error) {
	var (
		// the docker client is built by dive so that docker contexts are honored (see newDockerClient)
		dockerDaemon = docker.NewAPIClientProvider(Docker, tmp, ref, platform, newDockerClient)
		podmanDaemon = podman.NewDaemonProvider(tmp, ref, platform)
		// a stub that always errors unless built with the containers_image_openpgp tag
		store         = containerstorage.NewProvider(tmp, ref, platform)
		dockerArchive = docker.NewArchiveProvider(tmp, ref)
		ociArchive    = oci.NewArchiveProviderWithPlatform(tmp, ref, platform)
		ociDir        = oci.NewDirectoryProviderWithPlatform(tmp, ref, platform)
		// zero registry options mean credentials come from the docker config keychain
		registry         = oci.NewRegistryProvider(tmp, image.RegistryOptions{}, ref, platform)
		containerdDaemon = containerd.NewDaemonProvider(tmp, image.RegistryOptions{}, os.Getenv("CONTAINERD_NAMESPACE"), ref, platform)
	)

	switch src {
	case Docker:
		return []image.Provider{dockerDaemon}, nil
	case Podman:
		// on linux the local store is readable without a socket. Elsewhere podman runs in a VM
		if goos == "linux" {
			return []image.Provider{store, podmanDaemon}, nil
		}
		return []image.Provider{podmanDaemon}, nil
	case DockerArchive:
		return []image.Provider{dockerArchive}, nil
	case OCIArchive:
		return []image.Provider{ociArchive}, nil
	case OCIDir:
		return []image.Provider{ociDir}, nil
	case Registry:
		return []image.Provider{registry}, nil
	case Containerd:
		return []image.Provider{containerdDaemon}, nil
	case ContainersStorage:
		return []image.Provider{store}, nil
	case Auto:
		// same order as stereoscope's own detection, with dive's docker client
		return []image.Provider{dockerArchive, ociArchive, ociDir, dockerDaemon, podmanDaemon, containerdDaemon, store, registry}, nil
	}
	return nil, fmt.Errorf("unknown image source %q (valid sources: %s)", src, strings.Join(Sources, ", "))
}
