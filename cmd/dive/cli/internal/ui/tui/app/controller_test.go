package app

import (
	"context"
	"testing"

	"github.com/awesome-gocui/gocui"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/internal/imagefixture"
)

// newTestApp lays out the app on a simulated screen without running the main loop
func newTestApp(t *testing.T) *app {
	analysis, err := dive.Analyze(context.Background(), "test-docker-image", imagefixture.Load(t, "test-docker-image"))
	require.NoError(t, err)

	opts := options.DefaultApplication()
	require.NoError(t, opts.UI.Keybinding.PostLoad())

	g, err := gocui.NewGui(gocui.OutputSimulator, true)
	require.NoError(t, err)
	t.Cleanup(g.Close)

	a, err := newApp(g, tui.Config{Analysis: analysis, Preferences: opts.TUIPreferences()})
	require.NoError(t, err)
	require.NoError(t, a.layout.Layout(g))
	return a
}

func TestController_filterFocus(t *testing.T) {
	a := newTestApp(t)
	c := a.controller

	// focus the tree, then open the filter from there
	require.NoError(t, c.ToggleView())
	require.Equal(t, c.views.Tree.Name(), a.gui.CurrentView().Name())

	require.NoError(t, c.ToggleFilterView())
	require.Equal(t, c.views.Filter.Name(), a.gui.CurrentView().Name())
	require.Same(t, c.views.Filter, c.views.Status.CurrentView())

	// closing goes back to the tree, not the layer view
	require.NoError(t, c.CloseFilterView())
	require.Equal(t, c.views.Tree.Name(), a.gui.CurrentView().Name())
	require.Same(t, c.views.Tree, c.views.Status.CurrentView())
}
