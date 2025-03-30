package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultDockerfileName    = "Dockerfile"
	defaultContainerfileName = "Containerfile"
)

func buildImageFromCli(buildArgs []string) (string, error) {
	iidfile, err := os.CreateTemp("/tmp", "dive.*.iid")
	if err != nil {
		return "", err
	}
	defer os.Remove(iidfile.Name())
	defer iidfile.Close()

	var allArgs []string
	if isFileFlagsAreSet(buildArgs, "-f", "--file") {
		allArgs = append([]string{"--iidfile", iidfile.Name()}, buildArgs...)
	} else {
		containerFilePath, err := tryFindContainerfile(buildArgs)
		if err != nil {
			return "", err
		}
		allArgs = append([]string{"--iidfile", iidfile.Name(), "-f", containerFilePath}, buildArgs...)
	}

	err = runDockerCmd("build", allArgs...)
	if err != nil {
		return "", err
	}

	imageId, err := os.ReadFile(iidfile.Name())
	if err != nil {
		return "", err
	}

	return string(imageId), nil
}

// Checks if specified flags are present in the arguments list.
func isFileFlagsAreSet(args []string, flags ...string) bool {
	for i, arg := range args {
		for _, flag := range flags {
			if arg == flag && i+1 < len(args) {
				return true
			}
		}
	}
	return false
}

// This functions loops through a provided build arguments and tries to find a Containerfile or a Dockerfile.
func tryFindContainerfile(buildArgs []string) (string, error) {
	// Look for a build context within the provided build arguments.
	// Test build arguments one by one to find a valid path containing default names of `Containerfile` or a `Dockerfile` (in that order).
	for _, arg := range buildArgs {
		if fileInfo, err := os.Stat(arg); err == nil && fileInfo.IsDir() {
			containerfilePath := filepath.Join(arg, defaultContainerfileName)
			if _, err := os.Stat(containerfilePath); err == nil {
				return containerfilePath, nil
			}

			altContainerfilePath := filepath.Join(arg, strings.ToLower(defaultContainerfileName))
			if _, err := os.Stat(altContainerfilePath); err == nil {
				return altContainerfilePath, nil
			}

			dockerfilePath := filepath.Join(arg, defaultDockerfileName)
			if _, err := os.Stat(dockerfilePath); err == nil {
				return dockerfilePath, nil
			}

			altDockerfilePath := filepath.Join(arg, strings.ToLower(defaultDockerfileName))
			if _, err := os.Stat(altDockerfilePath); err == nil {
				return altDockerfilePath, nil
			}
		}
	}

	return "", fmt.Errorf("could not find Containerfile or Dockerfile\n")
}
