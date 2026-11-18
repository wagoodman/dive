package viewmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wagoodman/dive/dive"
)

func TestLayerStringIsSingleLine(t *testing.T) {
	for _, cmd := range []string{"a\nb", "a\r\nb", "a\rb"} {
		l := NewLayers([]dive.Layer{{Index: 1, Command: cmd, Size: 1}})[0]
		assert.Equal(t, "    1 B  a↵b", l.String(), "command %q", cmd)
	}
}
