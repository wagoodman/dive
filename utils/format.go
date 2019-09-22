package utils

import (
	"github.com/logrusorgru/aurora"
)

func TitleFormat(s string) string {
	return aurora.Bold(s).String()
}
