// Package source resolves a user provided image reference to a stereoscope image.
package source

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/anchore/stereoscope/pkg/image/containerd"
	"github.com/anchore/stereoscope/pkg/image/containerstorage"
	"github.com/anchore/stereoscope/pkg/image/docker"
	"github.com/anchore/stereoscope/pkg/image/oci"
	"github.com/anchore/stereoscope/pkg/image/podman"

	"github.com/wagoodman/dive/internal/bus"
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
	// Auto tries every source in stereoscope's detection order: files (so a file or dir in the working directory
	// named like the reference wins), then daemons and stores, then a registry pull, which means an image no local
	// daemon has can quietly come from Docker Hub. The source it lands on is reported. It is only valid as a --source
	// value, never as a reference scheme.
	Auto = "auto"
)

// Sources lists every valid --source value (aliases excluded).
var Sources = []string{Docker, Podman, DockerArchive, OCIArchive, OCIDir, Registry, Containerd, ContainersStorage, Auto}

var aliases = map[string]string{
	"docker-tar":   DockerArchive,
	"oci-registry": Registry,
}

// sources whose reference is a path, the only ones "<source>:<ref>" (no "//") selects. Any other source name followed
// by a colon reads as an image name and tag ("docker:dind", "registry:2.8")
var pathSources = []string{DockerArchive, OCIArchive, OCIDir}

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

// Parse splits a user reference into a source and the reference to hand to that source. "<source>://<ref>" works
// for every source, stereoscope's "<source>:<ref>" only for the file based ones (e.g. "oci-dir:./img"). Without a
// known scheme the (normalized) defaultSource is returned with the input untouched; an empty defaultSource means
// Docker.
func Parse(input, defaultSource string) (src, ref string) {
	if scheme, rest, ok := strings.Cut(input, "://"); ok {
		if s, ok := schemeSource(scheme); ok {
			return s, rest
		}
	}
	if scheme, rest, ok := strings.Cut(input, ":"); ok {
		if s, ok := schemeSource(scheme); ok && slices.Contains(pathSources, s) {
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

// tempDirs parents the temp dirs of every Load, so Cleanup can remove them even when the caller never got the
// image back (e.g. an interrupt while a provider is still copying the image).
var tempDirs = file.NewTempDirGenerator("dive")

// loading counts the Loads still running, Cleanup waits for them
var loading sync.WaitGroup

// Cleanup removes the temp dirs of every image Load has read (or is still reading). Loads still running get up to
// wait to return first, since a provider that is still copying would otherwise start a fresh temp dir after the
// cleanup. It is safe to call more than once and concurrently with img.Cleanup(). Meant for the end of the process,
// when interrupted work never returned.
func Cleanup(wait time.Duration) error {
	done := make(chan struct{})
	go func() {
		loading.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(wait):
	}
	return tempDirs.Cleanup()
}

// Load resolves the user reference (see Parse) and reads the image with stereoscope. The caller owns the returned
// image and should call img.Cleanup() when done with it; Cleanup is the backstop for paths that never get there.
func Load(ctx context.Context, input string, cfg Config) (*image.Image, error) {
	loading.Add(1)
	defer loading.Done()

	src, ref := Parse(input, cfg.Source)

	var platform *image.Platform
	if cfg.Platform != "" {
		p, err := image.NewPlatform(cfg.Platform)
		if err != nil {
			return nil, err
		}
		platform = p
	}

	tmp := tempDirs.NewGenerator()
	ps, err := providers(src, ref, runtime.GOOS, platform, tmp)
	if err != nil {
		return nil, err
	}

	var errs []error
	for _, p := range ps {
		if ctx.Err() != nil {
			errs = append(errs, ctx.Err())
			break
		}
		img, err := p.Provide(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", p.Name(), err))
			continue
		}
		if src == Auto {
			bus.Notify(fmt.Sprintf("auto detected the image source: %s", p.Name()))
		}
		if platform != nil && !platformMatches(img.Metadata, platform) {
			bus.Notify(fmt.Sprintf("the %s source has no %s image, using %s", p.Name(), platform, metadataPlatform(img.Metadata)))
		}
		return img, nil
	}
	// providers that failed may have left temp dirs behind
	errs = append(errs, tmp.Cleanup())
	return nil, fmt.Errorf("unable to load image %q from %s: %w", ref, src, errors.Join(errs...))
}

// platformMatches reports whether the image is the requested platform. Some providers (docker-archive) can't select
// a platform and read whatever the archive holds.
func platformMatches(m image.Metadata, p *image.Platform) bool {
	return m.OS == p.OS && m.Architecture == p.Architecture && (p.Variant == "" || m.Variant == p.Variant)
}

func metadataPlatform(m image.Metadata) string {
	s := m.OS + "/" + m.Architecture
	if m.Variant != "" {
		s += "/" + m.Variant
	}
	return s
}

// plainHTTPRegistry reports an error for a registry the registry client would reach over plain http. It falls back
// to http for private IP registries too, which would send docker config credentials in the clear and accept whatever
// image answers. Only "localhost" and loopback IP literals are allowed, nothing leaves the machine there. Other
// "*.localhost" names are refused: release builds use the pure go resolver, which sends them to /etc/hosts and DNS.
func plainHTTPRegistry(ref string) error {
	r, err := name.ParseReference(ref)
	if err != nil {
		// the registry provider reports the bad reference itself
		return nil
	}
	reg := r.Context().Registry
	if reg.Scheme() != "http" {
		return nil
	}
	host := reg.RegistryStr()
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	// a bare "[::1]" has no port for SplitHostPort to strip the brackets with
	host = strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
	if ip := net.ParseIP(host); host == "localhost" || (ip != nil && ip.IsLoopback()) {
		return nil
	}
	return fmt.Errorf("refusing to pull %q: the registry client would fall back to plain http for %s", ref, reg.RegistryStr())
}

// localName is the "localhost/" name podman gives a short name it built or tagged locally ("podman build -t myapp"
// stores localhost/myapp), which containers-storage won't find from the short name: it reads that as
// docker.io/library/myapp. It reports false for a reference that already names a registry.
func localName(ref string) (string, bool) {
	first, _, hasPath := strings.Cut(ref, "/")
	if hasPath && (first == "localhost" || strings.ContainsAny(first, ".:")) {
		return "", false
	}
	return "localhost/" + ref, true
}

// providers returns the stereoscope providers to try, in order, for the given source. Constructing them is cheap,
// nothing connects until Provide.
func providers(src, ref, goos string, platform *image.Platform, tmp *file.TempDirGenerator) ([]image.Provider, error) {
	var (
		// the docker client is built by dive so that docker contexts are honored (see newDockerClient)
		dockerDaemon = docker.NewAPIClientProvider(Docker, tmp, ref, platform, newDockerClient)
		// same as podman.NewDaemonProvider, without its ssh fallback (see newPodmanClient)
		podmanDaemon = docker.NewAPIClientProvider(string(podman.Daemon), tmp, ref, platform, newPodmanClient)
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
			ps := []image.Provider{store}
			if local, ok := localName(ref); ok {
				ps = append(ps, containerstorage.NewProvider(tmp, local, platform))
			}
			return append(ps, podmanDaemon), nil
		}
		return []image.Provider{podmanDaemon}, nil
	case DockerArchive:
		// main's own parser read OCI archives as docker-archive too
		return []image.Provider{dockerArchive, ociArchive}, nil
	case OCIArchive:
		return []image.Provider{ociArchive}, nil
	case OCIDir:
		return []image.Provider{ociDir}, nil
	case Registry:
		if err := plainHTTPRegistry(ref); err != nil {
			return nil, err
		}
		return []image.Provider{registry}, nil
	case Containerd:
		return []image.Provider{containerdDaemon}, nil
	case ContainersStorage:
		return []image.Provider{store}, nil
	case Auto:
		// same order as stereoscope's own detection, with dive's docker client
		ps := []image.Provider{dockerArchive, ociArchive, ociDir, dockerDaemon, podmanDaemon, containerdDaemon, store}
		if plainHTTPRegistry(ref) == nil {
			ps = append(ps, registry)
		}
		return ps, nil
	}
	return nil, fmt.Errorf("unknown image source %q (valid sources: %s)", src, strings.Join(Sources, ", "))
}
