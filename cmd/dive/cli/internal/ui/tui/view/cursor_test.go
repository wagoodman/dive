package view

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCursorStep_scrollsThroughContent(t *testing.T) {
	g := newTestGui(t)
	// 4 visible rows
	v, _ := g.SetView("details", 0, 0, 40, 5, 0)
	v.Wrap = true

	lines := make([]string, 10)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i)
	}
	_, err := fmt.Fprintln(v, strings.Join(lines, "\n"))
	require.NoError(t, err)

	selected := func() string {
		_, oy := v.Origin()
		_, cy := v.Cursor()
		require.Less(t, cy, 4, "cursor left the view")
		return v.ViewBufferLines()[oy+cy]
	}

	for range 6 {
		require.NoError(t, CursorDown(g, v))
	}
	require.Equal(t, "line 6", selected())

	// stops on the last line (not the blank line after it)
	for range 10 {
		require.NoError(t, CursorDown(g, v))
	}
	require.Equal(t, "line 9", selected())

	// paging is clamped to what's left
	require.NoError(t, CursorStep(g, v, -4))
	require.Equal(t, "line 5", selected())
	require.NoError(t, CursorStep(g, v, -4))
	require.Equal(t, "line 1", selected())
	require.NoError(t, CursorStep(g, v, -4))
	require.Equal(t, "line 0", selected())
	require.NoError(t, CursorStep(g, v, 20))
	require.Equal(t, "line 9", selected())
}

func TestCursorStep_countsWrappedLines(t *testing.T) {
	g := newTestGui(t)
	// 2 visible rows, 10 columns
	v, _ := g.SetView("details", 0, 0, 11, 3, 0)
	v.Wrap = true

	_, err := fmt.Fprintln(v, "0123456789abcdefghij\nlast")
	require.NoError(t, err)

	require.NoError(t, CursorStep(g, v, 5))
	_, oy := v.Origin()
	_, cy := v.Cursor()
	require.Equal(t, 1, oy)
	require.Equal(t, 1, cy)
	require.Equal(t, "last", v.ViewBufferLines()[oy+cy])
}
