package format

import (
	"strings"
	"testing"

	"github.com/lunixbochs/vtclean"
	"github.com/mattn/go-runewidth"
	"github.com/stretchr/testify/require"
)

func TestRenderHeader_fillsWidth(t *testing.T) {
	for _, selected := range []bool{true, false} {
		header := vtclean.Clean(strings.TrimSuffix(RenderHeader("Layers", 40, selected), "\n"), false)
		require.Equal(t, 40, runewidth.StringWidth(header), "selected=%v: %q", selected, header)
	}
}
