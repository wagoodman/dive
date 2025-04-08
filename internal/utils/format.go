package utils

import (
	"strings"
)

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
