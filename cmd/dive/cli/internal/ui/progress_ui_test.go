package ui

import (
	"bytes"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wagoodman/go-partybus"

	"github.com/anchore/bubbly"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/internal/bus/event"
	"github.com/wagoodman/dive/internal/bus/event/payload"
)

func TestNew_fallsBackWithoutTTY(t *testing.T) {
	// go test never gives the test binary a terminal on stdin/stderr
	_, ok := New(tui.DefaultPreferences(), &bytes.Buffer{}, false, 0, false).(*LineUI)
	assert.True(t, ok, "expected the line based UI without a TTY")

	t.Setenv("CI", "true")
	_, ok = New(tui.DefaultPreferences(), &bytes.Buffer{}, false, 0, false).(*LineUI)
	assert.True(t, ok, "expected the line based UI with CI set")
}

func TestWriteEvents_reportMatchesLineUI(t *testing.T) {
	report := "Analysis:\n  efficiency: 98.44 %\n\nFAIL [pass:1 fail:1 skip:1]\n"
	e := partybus.Event{Type: event.Report, Value: report}

	var lineOut bytes.Buffer
	require.NoError(t, NewLineUI(tui.DefaultPreferences(), &lineOut, false, 1).Handle(e))

	var progressOut bytes.Buffer
	require.NoError(t, writeEvents(&progressOut, &bytes.Buffer{}, e))

	assert.Equal(t, lineOut.String(), progressOut.String())
}

func TestWriteEvents_notificationsGoToStderr(t *testing.T) {
	var out, errOut bytes.Buffer
	require.NoError(t, writeEvents(&out, &errOut, partybus.Event{Type: event.Notification, Value: "heads up"}))
	assert.Empty(t, out.String())
	assert.Contains(t, errOut.String(), "heads up")
}

// startProgressUI runs the progress UI headless: no input, output to a buffer that is only read after teardown
func startProgressUI(t *testing.T, handlers ...bubbly.EventHandler) (*progressUI, *bytes.Buffer) {
	t.Helper()
	var errOut bytes.Buffer
	m := newProgressUI(tui.DefaultPreferences(), &bytes.Buffer{}, handlers...)
	m.err = &errOut
	m.in = nil
	require.NoError(t, m.Setup(nil))
	return m, &errOut
}

const slowEvent partybus.EventType = "test-slow-row"

// slowRow starts its row (adds to the wait group) only after a delay, then finishes it a bit later
type slowRow struct {
	wg       sync.WaitGroup
	finished atomic.Bool
}

func (h *slowRow) RespondsTo() []partybus.EventType { return []partybus.EventType{slowEvent} }

func (h *slowRow) Handle(partybus.Event) ([]tea.Model, tea.Cmd) {
	time.Sleep(50 * time.Millisecond)
	h.wg.Add(1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		h.finished.Store(true)
		h.wg.Done()
	}()
	return nil, nil
}

func (h *slowRow) Wait() { h.wg.Wait() }

func TestProgressUI_stopWaitsForRowsSentBeforeIt(t *testing.T) {
	h := &slowRow{}
	m, _ := startProgressUI(t, h)

	require.NoError(t, m.Handle(partybus.Event{Type: slowEvent}))
	// without the flush the wait group is still empty here, so stop wouldn't wait for the row
	m.stop(time.Second)
	assert.True(t, h.finished.Load(), "stop returned before the row finished")
	require.NoError(t, m.Teardown(false))
}

func TestProgressUI_interruptSkipsTUI(t *testing.T) {
	m, errOut := startProgressUI(t)

	m.program.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	require.Eventually(t, m.interrupted.Load, time.Second, 10*time.Millisecond)

	// the explore event queued behind the ctrl-c releases the publisher without starting gocui (which would need a
	// terminal here)
	var released bool
	require.NoError(t, m.Handle(partybus.Event{
		Type:  event.ExploreAnalysis,
		Value: payload.Explore{Analysis: &dive.Analysis{}, Done: func() { released = true }},
	}))
	assert.True(t, released)
	assert.True(t, m.stopped)

	// notifications published after the handoff still print on teardown
	require.NoError(t, m.Handle(partybus.Event{Type: event.Notification, Value: "extract of /etc: boom"}))
	require.NoError(t, m.Teardown(false))
	assert.Contains(t, errOut.String(), "extract of /etc: boom")
}

func TestProgressUI_notificationPrintsWhileRunning(t *testing.T) {
	m, errOut := startProgressUI(t)

	// a hostile image path trying to set the terminal title
	require.NoError(t, m.Handle(partybus.Event{Type: event.Notification, Value: "auto detected \x1b]0;pwned\a"}))
	m.eventsLock.Lock()
	assert.Empty(t, m.finalizeEvents, "printed above the frame, not held for teardown")
	m.eventsLock.Unlock()

	require.NoError(t, m.Teardown(false))
	out := errOut.String()
	assert.Equal(t, 1, strings.Count(out, "auto detected"))
	assert.Contains(t, out, `\x1b]0;pwned\a`)
	assert.NotContains(t, out, "\x1b]0;pwned")
}

func TestDrain(t *testing.T) {
	b := partybus.NewBus()
	sub := b.Subscribe()
	var released bool
	b.Publish(partybus.Event{Type: event.TaskStarted})
	b.Publish(partybus.Event{Type: event.Notification, Value: "extract of /etc: boom"})
	b.Publish(partybus.Event{
		Type:  event.ExploreAnalysis,
		Value: payload.Explore{Analysis: &dive.Analysis{}, Done: func() { released = true }},
	})

	events := drain(sub)
	require.Len(t, events, 1)
	assert.Equal(t, event.Notification, events[0].Type)
	assert.True(t, released, "a queued explore has to release its publisher")

	// clio already unsubscribed on a normal exit
	assert.Empty(t, drain(sub))
	assert.Empty(t, drain(nil))
	assert.Empty(t, drain((*partybus.Subscription)(nil)))
}
