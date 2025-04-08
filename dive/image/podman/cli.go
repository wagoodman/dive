//go:build linux || darwin

package podman

import (
	"fmt"
	"github.com/wagoodman/dive/internal/log"
	"github.com/wagoodman/dive/internal/utils"
	"io"
	"os"
	"os/exec"
	"strings"
)

// runPodmanCmd runs a given Podman command in the current tty
func runPodmanCmd(cmdStr string, args ...string) error {
	if !isPodmanClientBinaryAvailable() {
		return fmt.Errorf("cannot find podman client executable")
	}

	allArgs := utils.CleanArgs(append([]string{cmdStr}, args...))

	fullCmd := strings.Join(append([]string{"docker"}, allArgs...), " ")
	log.WithFields("cmd", fullCmd).Trace("executing")

	cmd := exec.Command("podman", allArgs...)
	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func streamPodmanCmd(args ...string) (error, io.Reader) {
	if !isPodmanClientBinaryAvailable() {
		return fmt.Errorf("cannot find podman client executable"), nil
	}

	allArgs := utils.CleanArgs(args)
	fullCmd := strings.Join(append([]string{"docker"}, allArgs...), " ")
	log.WithFields("cmd", fullCmd).Trace("executing (streaming)")

	cmd := exec.Command("podman", allArgs...)
	cmd.Env = os.Environ()

	reader, writer, err := os.Pipe()
	if err != nil {
		return err, nil
	}
	defer writer.Close()

	cmd.Stdout = writer
	cmd.Stderr = os.Stderr

	return cmd.Start(), reader
}

func isPodmanClientBinaryAvailable() bool {
	_, err := exec.LookPath("podman")
	return err == nil
}
