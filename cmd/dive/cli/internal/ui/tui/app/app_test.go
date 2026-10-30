package app

import (
	"context"
	"testing"
	"time"

	"github.com/awesome-gocui/gocui"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/internal/imagefixture"
)

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
