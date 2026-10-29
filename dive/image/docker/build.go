package docker

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/scylladb/go-set/strset"
	"github.com/spf13/afero"
)

const (
	defaultDockerfileName    = "Dockerfile"
	defaultContainerfileName = "Containerfile"
)

func buildImageFromCli(fs afero.Fs, buildArgs []string) (string, error) {
	iidfile, err := afero.TempFile(fs, "", "dive.*.iid")
	if err != nil {
		return "", err
	}
	defer fs.Remove(iidfile.Name()) // nolint:errcheck
	defer iidfile.Close()

	var allArgs []string
	if isFileFlagsAreSet(buildArgs, "-f", "--file") {
		allArgs = append([]string{"--iidfile", iidfile.Name()}, buildArgs...)
	} else {
		containerFilePath, err := tryFindContainerfile(fs, buildArgs)
		if err != nil {
			return "", err
		}
		allArgs = append([]string{"--iidfile", iidfile.Name(), "-f", containerFilePath}, buildArgs...)
	}

	err = runDockerCmd("build", allArgs...)
	if err != nil {
		return "", err
	}

	imageId, err := afero.ReadFile(fs, iidfile.Name())
	if err != nil {
		return "", err
	}

	return string(imageId), nil
}

// isFileFlagsAreSet Checks if specified flags are present in the argument list.
func isFileFlagsAreSet(args []string, flags ...string) bool {
	flagSet := strset.New(flags...)
	for i, arg := range args {
		if flagSet.Has(arg) && i+1 < len(args) {
			return true
		}
	}
	return false
}

// tryFindContainerfile loops through provided build arguments and tries to find a Containerfile or a Dockerfile.
func tryFindContainerfile(fs afero.Fs, buildArgs []string) (string, error) {
	// Look for a build context within the provided build arguments.
	// Test build arguments one by one to find a valid path containing default names of `Containerfile` or a `Dockerfile` (in that order).
	candidates := []string{
		defaultContainerfileName,                  // Containerfile
		strings.ToLower(defaultContainerfileName), // containerfile
		defaultDockerfileName,                     // Dockerfile
		strings.ToLower(defaultDockerfileName),    // dockerfile
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

	return "", fmt.Errorf("could not find Containerfile or Dockerfile\n")
}
