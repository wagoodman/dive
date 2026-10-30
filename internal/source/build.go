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
	"github.com/wagoodman/dive/internal/utils"
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
		// docker doesn't look for a Containerfile on its own
		if !isFileFlagsAreSet(args, "-f", "--file") {
			containerfile, err := tryFindContainerfile(fs, args)
			if err != nil {
				return "", err
			}
			allArgs = append(allArgs, "-f", containerfile)
		}
	case Podman:
	default:
		return "", fmt.Errorf("cannot build with source %q (supported: %s, %s)", engine, Docker, Podman)
	}
	allArgs = utils.CleanArgs(append(allArgs, args...))

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
	return Docker + "://" + id, nil
}

// isFileFlagsAreSet checks if any of the given flags is present in the argument list with a value after it.
func isFileFlagsAreSet(args []string, flags ...string) bool {
	for i, arg := range args {
		if slices.Contains(flags, arg) && i+1 < len(args) {
			return true
		}
	}
	return false
}

// tryFindContainerfile looks through the build args for a directory containing a Containerfile or a Dockerfile
// (in that order).
func tryFindContainerfile(fs afero.Fs, buildArgs []string) (string, error) {
	candidates := []string{
		defaultContainerfileName,
		strings.ToLower(defaultContainerfileName),
		defaultDockerfileName,
		strings.ToLower(defaultDockerfileName),
	}

	for _, arg := range buildArgs {
		fileInfo, err := fs.Stat(arg)
		if err == nil && fileInfo.IsDir() {
			for _, candidate := range candidates {
				filePath := filepath.Join(arg, candidate)
				if exists, _ := afero.Exists(fs, filePath); exists {
					return filePath, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not find Containerfile or Dockerfile")
}
