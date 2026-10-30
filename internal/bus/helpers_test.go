package bus

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wagoodman/go-partybus"

	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/internal/bus/event/payload"
)

func withBus(t *testing.T, grace time.Duration) *partybus.Bus {
	t.Helper()
	prevPublisher, prevGrace := publisher, exploreGrace
	t.Cleanup(func() { publisher, exploreGrace = prevPublisher, prevGrace })
	b := partybus.NewBus()
	Set(b)
	exploreGrace = grace
	return b
}

// explore runs ExploreAnalysis and fails the test if it hasn't returned within a few seconds
func explore(t *testing.T, ctx context.Context) {
	t.Helper()
	returned := make(chan struct{})
	go func() {
		ExploreAnalysis(ctx, &dive.Analysis{})
		close(returned)
	}()
	select {
	case <-returned:
	case <-time.After(5 * time.Second):
		t.Fatal("ExploreAnalysis did not return")
	}
}

// ui takes the explore event and calls done after the given delay, then reports through released
func ui(b *partybus.Bus, delay time.Duration, released *atomic.Bool) {
	sub := b.Subscribe()
	go func() {
		e := <-sub.Events()
		time.Sleep(delay)
		released.Store(true)
		e.Value.(payload.Explore).Done()
	}()
}

func TestExploreAnalysis_returnsOnDone(t *testing.T) {
	b := withBus(t, time.Hour)
	var released atomic.Bool
	ui(b, 0, &released)

	explore(t, context.Background())
	assert.True(t, released.Load())
}

func TestExploreAnalysis_cancelledWaitsForDone(t *testing.T) {
	// the TUI reads from the image until it calls done, the grace must not cut that short
	b := withBus(t, time.Hour)
	var released atomic.Bool
	ui(b, 100*time.Millisecond, &released)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	explore(t, ctx)
	assert.True(t, released.Load(), "returned before the UI called done")
}

func TestExploreAnalysis_cancelledNobodyListening(t *testing.T) {
	withBus(t, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	explore(t, ctx)
	assert.GreaterOrEqual(t, time.Since(start), 50*time.Millisecond)
}

func TestExploreAnalysis_noPublisher(t *testing.T) {
	withBus(t, time.Hour)
	Set(nil)

	// a ctx that is never done, only the missing publisher can make this return
	explore(t, context.Background())
	require.Nil(t, Get())
}
