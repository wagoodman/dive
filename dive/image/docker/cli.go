package docker

import (
	"fmt"
	"github.com/wagoodman/dive/internal/utils"
	"os"
	"os/exec"
)

// runDockerCmd runs a given Docker command in the current tty
func runDockerCmd(cmdStr string, args ...string) error {
	if !isDockerClientBinaryAvailable() {
		return fmt.Errorf("cannot find docker client executable")
	}

	allArgs := utils.CleanArgs(append([]string{cmdStr}, args...))

	cmd := exec.Command("docker", allArgs...)
	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func isDockerClientBinaryAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}
