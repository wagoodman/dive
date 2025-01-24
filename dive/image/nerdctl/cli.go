package nerdctl

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/wagoodman/dive/utils"
)

// runNerdctlCmd runs a given nerdctl command in the current tty
func runNerdctlCmd(cmdStr string, args ...string) error {
	if !isNerdctlBinaryAvailable() {
		return fmt.Errorf("cannot find nerdctl executable")
	}

	allArgs := utils.CleanArgs(append([]string{cmdStr}, args...))

	cmd := exec.Command("nerdctl", allArgs...)
	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func streamNerdctlCmd(args ...string) (io.Reader, error) {
	if !isNerdctlBinaryAvailable() {
		return nil, fmt.Errorf("cannot find nerdctl executable")
	}

	cmd := exec.Command("nerdctl", utils.CleanArgs(args)...)
	cmd.Env = os.Environ()

	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	defer writer.Close()

	cmd.Stdout = writer
	cmd.Stderr = os.Stderr

	return reader, cmd.Start()
}

func isNerdctlBinaryAvailable() bool {
	_, err := exec.LookPath("nerdctl")
	return err == nil
}
