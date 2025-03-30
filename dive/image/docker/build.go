package docker

import (
	"fmt"
	"os"
	"path/filepath"
)

func buildImageFromCli(buildArgs []string) (string, error) {
	iidfile, err := os.CreateTemp("/tmp", "dive.*.iid")
	if err != nil {
		return "", err
	}
	defer os.Remove(iidfile.Name())
	defer iidfile.Close()

	containerFilePath, err := tryFindContainerFile(buildArgs)
	if err != nil {
		return "", err
	}

	allArgs := append([]string{"--iidfile", iidfile.Name(), "-f", containerFilePath}, buildArgs...)
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

// This functions loops through a provided build arguments and tries to find a Containerfile or a Dockerfile.
func tryFindContainerFile(buildArgs []string) (string, error) {
	var fileFlag bool
	var filePath string

	// If the `-f` flag was set assume the correct path was provided and use it...
	for i, arg := range buildArgs {
		if arg == "-f" && i+1 < len(buildArgs) {
			fileFlag = true
			filePath = buildArgs[i+1]
			break
		}
	}

	if fileFlag {
		return filePath, nil
	}

	// ... otherwise we have to look for a build context within the provided build arguments.
	// Test build arguments one by one to find a valid path containing `Containerfile` or a `Dockerfile` (in that order).
	for _, arg := range buildArgs {
		if fileInfo, err := os.Stat(arg); err == nil && fileInfo.IsDir() {
			containerfilePath := filepath.Join(arg, "Containerfile")
			dockerfilePath := filepath.Join(arg, "Dockerfile")

			if _, err := os.Stat(containerfilePath); err == nil {
				return containerfilePath, nil
			}
			if _, err := os.Stat(dockerfilePath); err == nil {
				return dockerfilePath, nil
			}
		}
	}

	return "", fmt.Errorf("could not find Containerfile or Dockerfile\n")
}
