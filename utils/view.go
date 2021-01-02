package utils

import (
	"github.com/awesome-gocui/gocui"
	"github.com/sirupsen/logrus"
)

// isNewView determines if a view has already been created based on the set of errors given (a bit hokie)
func IsNewView(errs ...error) bool {
	for _, err := range errs {
		if err == nil {
			return false
		}
		if err != gocui.ErrUnknownView {
			logrus.Errorf("IsNewView() unexpected error: %+v", err)
			return true
		}
	}
	return true
}
