package view

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/dive"
)

func TestImageDetails_RenderWritesBodyOnce(t *testing.T) {
	g := newTestGui(t)
	header, _ := g.SetView("imageDetailsHeader", 0, 0, 60, 3, 0)
	body, _ := g.SetView("imageDetails", 0, 2, 60, 8, 0)
	body.Wrap = true

	inefficiencies := make([]dive.Inefficiency, 20)
	for i := range inefficiencies {
		inefficiencies[i] = dive.Inefficiency{Path: fmt.Sprintf("/file-%d", i), Count: 2, SizeBytes: 1000}
	}
	v := &ImageDetails{gui: g, body: body, header: header, imageName: "img", inefficiencies: inefficiencies}

	require.NoError(t, v.Render())
	lines := body.BufferLines()
	require.Contains(t, lines, "    2        1.0 kB  /file-19")
	require.Contains(t, body.Buffer(), "Potential wasted space: 20 kB")

	require.NoError(t, CursorStep(g, body, 12))
	_, oy := body.Origin()
	_, cy := body.Cursor()

	// moving between layers re-renders every pane, which must not rebuild the report or lose the scroll position
	require.NoError(t, v.Render())
	require.Equal(t, lines, body.BufferLines())
	_, oy2 := body.Origin()
	_, cy2 := body.Cursor()
	require.Equal(t, oy, oy2)
	require.Equal(t, cy, cy2)
}
