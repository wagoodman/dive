//go:build windows
// +build windows

package ui

import (
	"github.com/awesome-gocui/gocui"
)

// handle ctrl+z not supported on windows
func handle_ctrl_z(_ *gocui.Gui, _ *gocui.View) error {
	return nil
}
