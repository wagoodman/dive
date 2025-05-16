package ui

import (
	"github.com/muesli/termenv"
	"os"
	"strings"
)

var _ termenv.Environ = (*environWithoutCI)(nil)

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
