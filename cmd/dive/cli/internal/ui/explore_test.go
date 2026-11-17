package ui

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/anchore/go-logger"
	"github.com/anchore/go-logger/adapter/logrus"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wagoodman/go-partybus"

	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
	"github.com/wagoodman/dive/internal/bus/event"
	"github.com/wagoodman/dive/internal/bus/event/payload"
	"github.com/wagoodman/dive/internal/log"
)

func TestMuteConsoleLogs_restores(t *testing.T) {
	l, err := logrus.New(logrus.Config{EnableConsole: true, Level: logger.InfoLevel})
	require.NoError(t, err)
	var buf bytes.Buffer
	l.(logger.Controller).SetOutput(&buf)

	prev := log.Get()
	log.Set(l)
	t.Cleanup(func() { log.Set(prev) })

	restore := muteConsoleLogs()
	// a UI inside a UI (the TUI after the progress UI) restores in order
	restoreInner := muteConsoleLogs()
	log.Info("while the UI runs")
	restoreInner()
	log.Info("still muted")
	restore()
	assert.Empty(t, buf.String())

	// the logger clio reports the final error with is the same one, it has to print again
	l.Error("cannot load image")
	assert.Contains(t, buf.String(), "cannot load image")
}

func TestMuteConsoleLogs_keepsLogFile(t *testing.T) {
	l, err := logrus.New(logrus.Config{EnableConsole: true, Level: logger.InfoLevel})
	require.NoError(t, err)
	var console, file bytes.Buffer
	l.(logger.Controller).SetOutput(io.MultiWriter(&console, &file))

	prev, prevFile := log.Get(), log.File()
	log.Set(l)
	log.SetFile(&file)
	t.Cleanup(func() {
		log.Set(prev)
		log.SetFile(prevFile)
	})

	restore := muteConsoleLogs()
	log.Info("while the UI runs")
	restore()

	assert.Empty(t, console.String())
	assert.Contains(t, file.String(), "while the UI runs")
}

func TestLineUI_teardownRestoresLogs(t *testing.T) {
	l, err := logrus.New(logrus.Config{EnableConsole: true, Level: logger.InfoLevel})
	require.NoError(t, err)
	var buf bytes.Buffer
	l.(logger.Controller).SetOutput(&buf)

	prev := log.Get()
	log.Set(l)
	t.Cleanup(func() { log.Set(prev) })

	// default verbosity mutes logs while the UI prints its own status lines
	u := NewLineUI(tui.DefaultPreferences(), &bytes.Buffer{}, false, 0)
	require.NoError(t, u.Setup(nil))
	log.Info("duplicate of a status line")
	require.NoError(t, u.Teardown(false))
	assert.Empty(t, buf.String())

	// clio reports the command's error through this logger after teardown
	l.Error("cannot load image")
	assert.Contains(t, buf.String(), "cannot load image")
}

func TestLineUI_notifications(t *testing.T) {
	for _, verbosity := range []int{0, 1, 2} {
		var errOut bytes.Buffer
		u := NewLineUI(tui.DefaultPreferences(), &bytes.Buffer{}, false, verbosity)
		u.err = &errOut
		require.NoError(t, u.Handle(partybus.Event{Type: event.Notification, Value: "extract of /etc/\x1b[2Jx\r: boom\nsecond"}))
		// the logger never says this, so -v must not hide it
		assert.Contains(t, errOut.String(), "extract of /etc/\\x1b[2Jx\\r: boom\nsecond", "verbosity %d", verbosity)
		assert.NotContains(t, errOut.String(), "\x1b[2J")
	}

	var errOut bytes.Buffer
	u := NewLineUI(tui.DefaultPreferences(), &bytes.Buffer{}, true, 0)
	u.err = &errOut
	require.NoError(t, u.Handle(partybus.Event{Type: event.Notification, Value: "heads up"}))
	assert.Empty(t, errOut.String(), "--quiet still silences notifications")
}

func TestLineUI_teardownDrainsNotifications(t *testing.T) {
	// on SIGINT clio stops reading the bus, whatever is still queued is printed on teardown
	b := partybus.NewBus()
	var errOut bytes.Buffer
	u := NewLineUI(tui.DefaultPreferences(), &bytes.Buffer{}, false, 1)
	u.err = &errOut
	require.NoError(t, u.Setup(b.Subscribe()))
	b.Publish(partybus.Event{Type: event.Notification, Value: "extract of /etc: boom"})
	require.NoError(t, u.Teardown(true))
	assert.Contains(t, errOut.String(), "extract of /etc: boom")
}

func TestLineUI_taskStartedWithoutProgress(t *testing.T) {
	var errOut bytes.Buffer
	u := NewLineUI(tui.DefaultPreferences(), &bytes.Buffer{}, false, 0)
	u.err = &errOut

	// unparsable, and parsable but without a progress value
	require.NoError(t, u.Handle(partybus.Event{Type: event.TaskStarted, Source: "not a task"}))
	require.NoError(t, u.Handle(partybus.Event{Type: event.TaskStarted, Source: payload.GenericTask{Title: payload.Title{Default: "Loading image"}}}))
	assert.Contains(t, errOut.String(), "Loading image")
}

func TestLineUI_stderrPicksStatusColors(t *testing.T) {
	// stdout is a color terminal, stderr is not (dive --ci img 2>ci.log)
	prev := lipgloss.DefaultRenderer()
	t.Cleanup(func() { lipgloss.SetDefaultRenderer(prev) })
	stdout := lipgloss.NewRenderer(os.Stdout)
	stdout.SetColorProfile(termenv.ANSI256)
	lipgloss.SetDefaultRenderer(stdout)

	var errOut bytes.Buffer
	u := NewLineUI(tui.DefaultPreferences(), &bytes.Buffer{}, false, 0)
	u.err = &errOut
	require.NoError(t, u.Handle(partybus.Event{Type: event.Notification, Value: "heads up"}))
	require.NoError(t, u.Handle(partybus.Event{Type: event.TaskStarted, Source: payload.GenericTask{Title: payload.Title{Default: "Loading image"}, Context: "img"}}))
	assert.NotContains(t, errOut.String(), "\x1b[")
}

func TestLineUI_quietStillPrintsReport(t *testing.T) {
	var out, errOut bytes.Buffer
	u := NewLineUI(tui.DefaultPreferences(), &out, true, 0)
	u.err = &errOut
	require.NoError(t, u.Handle(partybus.Event{Type: event.Report, Value: "FAIL [pass:1 fail:1 skip:1]"}))
	assert.Equal(t, "FAIL [pass:1 fail:1 skip:1]\n", out.String())
	assert.Empty(t, errOut.String())
}
