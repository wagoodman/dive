package components

import (
	"fmt"
	"github.com/wagoodman/dive/dive/image"
	"strings"
)

func LayerDetailsText(layer *image.Layer) string {
	lines := []string{}
	if layer.Names != nil && len(layer.Names) > 0 {
		lines = append(lines, boldString("Tags:   ")+strings.Join(layer.Names, ", "))
	} else {
		lines = append(lines, boldString("Tags:   ")+"(none)")
	}
	lines = append(lines, boldString("Id:     ")+layer.Id)
	lines = append(lines, boldString("Digest: ")+layer.Digest)
	lines = append(lines, boldString("Command:"))
	lines = append(lines, layer.Command)
	return strings.Join(lines, "\n")
}

func boldString(s string) string {
	return fmt.Sprintf("[::b]%s[::-]", s)
}