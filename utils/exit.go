package utils

import (
	"github.com/k0kubun/go-ansi"
	"os"
)

// Note: this should only be used when exiting from non-gocui code
func Exit(rc int) {
	Cleanup()
	os.Exit(rc)
}

func Cleanup() {
	ansi.CursorShow()
}
