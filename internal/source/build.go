package source

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/afero"

	"github.com/wagoodman/dive/internal/log"
)

const (
	defaultDockerfileName    = "Dockerfile"
	defaultContainerfileName = "Containerfile"
)

// Build runs `docker build` or `podman build` (engine is Docker or Podman) with the user's build args, attached to
// the current tty. It returns a reference that Load resolves to the built image: the docker daemon for docker, and
// for podman containers-storage on linux with the podman socket as the fallback.
func Build(ctx context.Context, engine string, args []string) (string, error) {
	fs := afero.NewOsFs()

	iidfile, err := afero.TempFile(fs, "", "dive.*.iid")
	if err != nil {
		return "", err
	}
	defer fs.Remove(iidfile.Name()) // nolint:errcheck
	iidfile.Close()

	allArgs := []string{"build", "--iidfile", iidfile.Name()}
	switch engine {
	case Docker:
		// docker doesn't look for a Containerfile on its own. When none is found (or the context is stdin or a URL)
		// docker gets to decide
		if buildCtx, hasFile := parseBuildArgs(args); !hasFile {
			if containerfile := findContainerfile(fs, buildCtx); containerfile != "" {
				allArgs = append(allArgs, "-f", containerfile)
			}
		}
	case Podman:
	default:
		return "", fmt.Errorf("cannot build with source %q (supported: %s, %s)", engine, Docker, Podman)
	}
	// user args go through untouched: an empty or space padded value is meaningful to the build
	allArgs = append(allArgs, args...)

	log.WithFields("cmd", strings.Join(append([]string{engine}, allArgs...), " ")).Trace("executing")
	cmd := exec.CommandContext(ctx, engine, allArgs...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s build failed: %w", engine, err)
	}

	raw, err := afero.ReadFile(fs, iidfile.Name())
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(string(raw))
	if id == "" {
		return "", fmt.Errorf("%s build did not report an image ID", engine)
	}

	if engine == Podman {
		// containers-storage only resolves a bare hex ID, not "sha256:<hex>" (the podman socket takes either)
		return Podman + "://" + strings.TrimPrefix(id, "sha256:"), nil
	}

	// a buildx builder other than the docker driver keeps the image in its own cache unless told to --load, and
	// loading a missing ID from the daemon ends in a pull of "sha256:<id>" from Docker Hub
	if out, err := exec.CommandContext(ctx, Docker, "image", "inspect", "--format", "{{.Id}}", id).CombinedOutput(); err != nil {
		return "", fmt.Errorf("the built image %s is not in the docker daemon (add --load when building with a buildx builder that doesn't use the docker driver): %s", id, strings.TrimSpace(string(out)))
	}
	return Docker + "://" + id, nil
}

// docker build flags that take no value, every other flag consumes the next arg unless written as --flag=value or -fvalue
var buildBoolFlags = []string{
	"--check", "--compress", "--debug", "-D", "--disable-content-trust", "--force-rm", "--load", "--no-cache",
	"--pull", "--push", "-q", "--quiet", "--rm", "--squash",
}

// parseBuildArgs finds the build context (the last positional arg) in docker build args and whether a Dockerfile was
// given with -f/--file (in any of its spellings).
func parseBuildArgs(args []string) (buildCtx string, hasFile bool) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-" || !strings.HasPrefix(arg, "-"):
			buildCtx = arg
		case strings.HasPrefix(arg, "--"):
			name, _, inline := strings.Cut(arg, "=")
			hasFile = hasFile || name == "--file"
			if !inline && !slices.Contains(buildBoolFlags, name) {
				i++
			}
		default:
			hasFile = hasFile || strings.HasPrefix(arg, "-f")
			// a short flag longer than "-x" carries its value ("-fDockerfile", "-f=Dockerfile")
			if len(arg) == 2 && !slices.Contains(buildBoolFlags, arg) {
				i++
			}
		}
	}
	return buildCtx, hasFile
}

// findContainerfile returns the Containerfile or Dockerfile (in that order) in the build context directory, empty
// when the context isn't a local directory or has neither.
func findContainerfile(fs afero.Fs, buildCtx string) string {
	if info, err := fs.Stat(buildCtx); err != nil || !info.IsDir() {
		return ""
	}
	for _, candidate := range []string{
		defaultContainerfileName,
		strings.ToLower(defaultContainerfileName),
		defaultDockerfileName,
		strings.ToLower(defaultDockerfileName),
	} {
		p := filepath.Join(buildCtx, candidate)
		if exists, _ := afero.Exists(fs, p); exists {
			return p
		}
	}
	return ""
}
