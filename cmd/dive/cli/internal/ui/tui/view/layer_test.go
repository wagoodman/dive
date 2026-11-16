package view

import (
	"testing"

	"github.com/awesome-gocui/gocui"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
	"github.com/wagoodman/dive/dive"
)

// newTestGui creates a gui over gocui's simulated screen.
func newTestGui(t *testing.T) *gocui.Gui {
	t.Helper()
	g, err := gocui.NewGui(gocui.OutputSimulator, true)
	require.NoError(t, err)
	t.Cleanup(g.Close)
	return g
}

// newTestLayerView builds a layer pane whose body shows bodyRows rows.
func newTestLayerView(t *testing.T, layerCount, bodyRows int) *Layer {
	t.Helper()
	g := newTestGui(t)

	layers := make([]dive.Layer, layerCount)
	for i := range layers {
		layers[i] = dive.Layer{Index: i, Digest: "sha256:abc", Command: "RUN something"}
	}
	v, err := newLayerView(g, tui.Config{Analysis: &dive.Analysis{Layers: layers}, Preferences: tui.DefaultPreferences()})
	require.NoError(t, err)

	header, _ := g.SetView("layerHeader", 0, 0, 60, 3, 0)
	// a view shows y1-y0-1 rows
	body, _ := g.SetView("layer", 0, 2, 60, 2+bodyRows+1, 0)
	v.header, v.body = header, body
	return v
}

func TestLayer_PageDownStopsAtLastLayer(t *testing.T) {
	// from here one page lands exactly one past the last layer
	v := newTestLayerView(t, 14, 7)
	require.NoError(t, v.SetCursor(7))

	require.NoError(t, v.PageDown())
	require.Equal(t, 13, v.vm.LayerIndex)
	require.Equal(t, 13, v.CurrentLayer().Index)

	require.NoError(t, v.PageDown())
	require.Equal(t, 13, v.vm.LayerIndex)
}

func TestLayer_ScrollsOnlyWhenSelectionLeavesPane(t *testing.T) {
	// 30 layers in a 10 row pane
	v := newTestLayerView(t, 30, 10)

	for _, step := range []struct{ layer, origin int }{
		{layer: 25, origin: 16}, // selection pinned to the bottom row
		{layer: 20, origin: 16}, // moving up within the pane doesn't scroll
		{layer: 9, origin: 9},   // selection pinned to the top row
		{layer: 12, origin: 9},
		{layer: 0, origin: 0},
	} {
		require.NoError(t, v.SetCursor(step.layer))
		_, oy := v.body.Origin()
		require.Equal(t, step.origin, oy, "layer %d", step.layer)
	}
}

func TestLayer_RenderDrawsInOrder(t *testing.T) {
	// no main loop is running, so anything not drawn by Render itself never shows up
	v := newTestLayerView(t, 3, 10)

	_, err := v.gui.SetCurrentView(v.body.Name())
	require.NoError(t, err)
	require.NoError(t, v.Render())
	require.Contains(t, v.header.Buffer(), "● Layers")

	_, err = v.gui.SetCurrentView(v.header.Name())
	require.NoError(t, err)
	require.NoError(t, v.Render())
	require.NotContains(t, v.header.Buffer(), "●")
	require.Contains(t, v.header.Buffer(), "Layers")
}

func TestLayer_SetCursorClampsToLayers(t *testing.T) {
	v := newTestLayerView(t, 3, 10)

	require.NoError(t, v.SetCursor(3))
	require.Equal(t, 2, v.vm.LayerIndex)

	require.NoError(t, v.SetCursor(-1))
	require.Equal(t, 0, v.vm.LayerIndex)
}
