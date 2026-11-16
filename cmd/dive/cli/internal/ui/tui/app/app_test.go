package app

import (
	"context"
	"testing"
	"time"

	"github.com/awesome-gocui/gocui"
	"github.com/mattn/go-runewidth"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/internal/imagefixture"
)

func TestNarrowAmbiguousWidth(t *testing.T) {
	orig := runewidth.DefaultCondition.EastAsianWidth
	t.Cleanup(func() { runewidth.DefaultCondition.EastAsianWidth = orig })

	// as go-runewidth sets it up under a CJK locale
	runewidth.DefaultCondition.EastAsianWidth = true
	t.Setenv("RUNEWIDTH_EASTASIAN", "")
	narrowAmbiguousWidth()
	require.Equal(t, 1, runewidth.RuneWidth('│'))

	// an explicit choice is kept
	runewidth.DefaultCondition.EastAsianWidth = true
	t.Setenv("RUNEWIDTH_EASTASIAN", "1")
	narrowAmbiguousWidth()
	require.Equal(t, 2, runewidth.RuneWidth('│'))
}

func TestRun_noTerminalSize(t *testing.T) {
	orig := termSize
	t.Cleanup(func() { termSize = orig })
	termSize = func() (int, int, error) { return 0, 0, nil }

	// returns before gocui takes the terminal (which would wait for a resize)
	err := Run(context.Background(), tui.Config{Analysis: &dive.Analysis{Layers: make([]dive.Layer, 1)}})
	require.ErrorContains(t, err, "the terminal has no size (0x0)")
}

func TestRun_noLayers(t *testing.T) {
	outputMode = gocui.OutputSimulator
	t.Cleanup(func() { outputMode = gocui.OutputNormal })

	img := imagefixture.LoadSpec(t, imagefixture.Spec{})
	analysis, err := dive.Analyze(context.Background(), "empty", img)
	require.NoError(t, err)

	opts := options.DefaultApplication()
	require.NoError(t, opts.UI.Keybinding.PostLoad())

	err = Run(context.Background(), tui.Config{Analysis: analysis, Preferences: opts.TUIPreferences()})
	require.EqualError(t, err, "image has no layers")
}

func TestRun_quitsOnCancel(t *testing.T) {
	outputMode = gocui.OutputSimulator
	t.Cleanup(func() { outputMode = gocui.OutputNormal })

	analysis, err := dive.Analyze(context.Background(), "test-docker-image", imagefixture.Load(t, "test-docker-image"))
	require.NoError(t, err)

	// the key configs are only parsed when the config loads
	opts := options.DefaultApplication()
	require.NoError(t, opts.UI.Keybinding.PostLoad())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errs := make(chan error, 1)
	go func() {
		errs <- Run(ctx, tui.Config{Analysis: analysis, Preferences: opts.TUIPreferences()})
	}()

	select {
	case err := <-errs:
		t.Fatalf("the TUI quit on its own: %v", err)
	case <-time.After(500 * time.Millisecond):
	}

	// a SIGINT while the TUI runs (the terminal is raw, so it can only come from outside) cancels ctx
	cancel()
	select {
	case err := <-errs:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("the TUI did not quit on cancel")
	}
}
