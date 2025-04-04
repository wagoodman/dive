package utils

import (
	"errors"
	"github.com/awesome-gocui/gocui"
	"github.com/wagoodman/dive/internal/log"
)

// IsNewView determines if a view has already been created based on the set of errors given (a bit hokie)
func IsNewView(errs ...error) bool {
	for _, err := range errs {
		if err == nil {
			return false
		}
		if !errors.Is(err, gocui.ErrUnknownView) {
			log.WithFields("error", err).Error("IsNewView() unexpected error")
			return true
		}
	}
	return true
}
