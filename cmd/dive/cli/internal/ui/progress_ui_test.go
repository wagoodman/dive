package ui

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wagoodman/go-partybus"

	v1 "github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1"
	"github.com/wagoodman/dive/internal/bus/event"
)

func TestNew_fallsBackWithoutTTY(t *testing.T) {
	// go test never gives the test binary a terminal on stdin/stderr
	_, ok := New(v1.DefaultPreferences(), &bytes.Buffer{}, false, 0).(*V1UI)
	assert.True(t, ok, "expected the line based UI without a TTY")

	t.Setenv("CI", "true")
	_, ok = New(v1.DefaultPreferences(), &bytes.Buffer{}, false, 0).(*V1UI)
	assert.True(t, ok, "expected the line based UI with CI set")
}

func TestWriteEvents_reportMatchesLineUI(t *testing.T) {
	report := "Analysis:\n  efficiency: 98.44 %\n\nFAIL [pass:1 fail:1 skip:1]\n"
	e := partybus.Event{Type: event.Report, Value: report}

	var lineOut bytes.Buffer
	require.NoError(t, NewV1UI(v1.DefaultPreferences(), &lineOut, false, 1).Handle(e))

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
