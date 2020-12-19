package utils

import (
	"strings"

	"github.com/logrusorgru/aurora"
)

func TitleFormat(s string) string {
	return aurora.Bold(s).String()
}

// CleanArgs trims the whitespace from the given set of strings.
func CleanArgs(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, strings.Trim(str, " "))
		}
	}
	return r
}
