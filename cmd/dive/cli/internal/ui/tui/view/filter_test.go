package view

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilter_RenderKeepsOneLabel(t *testing.T) {
	g := newTestGui(t)
	v := newFilterView(g)
	v.header, _ = g.SetView("filterlabel", 0, 0, 20, 2, 0)

	for range 3 {
		require.NoError(t, v.Render())
	}
	require.Equal(t, "Path Filter: \n", v.header.Buffer())
}
