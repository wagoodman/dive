//go:build !windows
// +build !windows

package ui

import (
	"syscall"

	"github.com/awesome-gocui/gocui"
)

// handle ctrl+z
func handle_ctrl_z(g *gocui.Gui, v *gocui.View) error {
	gocui.Suspend()
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGSTOP); err != nil {
		return err
	}
	return gocui.Resume()
}
