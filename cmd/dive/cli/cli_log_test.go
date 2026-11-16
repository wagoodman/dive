package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anchore/clio"
	"github.com/anchore/go-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/internal/log"
)

func Test_newLogger_recordsLogFile(t *testing.T) {
	t.Cleanup(func() { log.SetFile(nil) })
	path := filepath.Join(t.TempDir(), "dive.log")

	l, err := newLogger(clio.Config{Log: &clio.LoggingConfig{FileLocation: path, Level: logger.InfoLevel}}, nil)
	require.NoError(t, err)
	require.NotNil(t, log.File(), "a UI muting the console would lose the log file too")

	// what a UI does while it draws
	l.(logger.Controller).SetOutput(log.File())
	l.Info("while the UI runs")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "while the UI runs")
}
