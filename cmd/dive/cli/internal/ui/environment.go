package ui

import (
	"os"
	"strings"

	"github.com/muesli/termenv"
)

var _ termenv.Environ = (*environWithoutCI)(nil)

// environWithoutCI hides the CI env var from termenv, which would otherwise turn off colors whenever CI is set
type environWithoutCI struct {
}

func (e environWithoutCI) Environ() []string {
	var out []string
	for _, s := range os.Environ() {
		if strings.HasPrefix(s, "CI=") {
			continue
		}
		out = append(out, s)
	}
	return out
}

func (e environWithoutCI) Getenv(s string) string {
	if s == "CI" {
		return ""
	}
	return os.Getenv(s)
}
