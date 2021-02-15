package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/logger"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// LogrusConfig contains all configurable values for the Logrus logger
type LogrusConfig struct {
	EnableConsole bool
	EnableFile    bool
	Structured    bool
	Level         string
	FileLocation  string
}

// LogrusLogger contains all runtime values for using Logrus with the configured output target and input configuration values.
type LogrusLogger struct {
	Config LogrusConfig
	Logger *logrus.Logger
	Output io.Writer
}

// NewLogrusLogger creates a new LogrusLogger with the given configuration
func NewLogrusLogger(cfg LogrusConfig) *LogrusLogger {
	appLogger := logrus.New()

	var output io.Writer
	switch {
	case cfg.EnableConsole && cfg.EnableFile:
		logFile, err := os.OpenFile(cfg.FileLocation, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			panic(fmt.Errorf("unable to setup log file: %w", err))
		}
		output = io.MultiWriter(os.Stderr, logFile)
	case cfg.EnableConsole:
		output = os.Stderr
	case cfg.EnableFile:
		logFile, err := os.OpenFile(cfg.FileLocation, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			panic(fmt.Errorf("unable to setup log file: %w", err))
		}
		output = logFile
	default:
		output = ioutil.Discard
	}

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		panic(err)
	}

	appLogger.SetOutput(output)
	appLogger.SetLevel(level)

	if cfg.Structured {
		appLogger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat:   "2006-01-02 15:04:05",
			DisableTimestamp:  false,
			DisableHTMLEscape: false,
			PrettyPrint:       false,
		})
	} else {
		appLogger.SetFormatter(&prefixed.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
			ForceFormatting: true,
		})
	}

	return &LogrusLogger{
		Config: cfg,
		Logger: appLogger,
		Output: output,
	}
}

// Tracef takes a formatted template string and template arguments for the trace logging level.
func (l *LogrusLogger) Tracef(format string, args ...interface{}) {
	l.Logger.Tracef(format, args...)
}

// Debugf takes a formatted template string and template arguments for the debug logging level.
func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.Logger.Debugf(format, args...)
}

// Infof takes a formatted template string and template arguments for the info logging level.
func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.Logger.Infof(format, args...)
}

// Warnf takes a formatted template string and template arguments for the warning logging level.
func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	l.Logger.Warnf(format, args...)
}

// Errorf takes a formatted template string and template arguments for the error logging level.
func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
}

// Trace logs the given arguments at the trace logging level.
func (l *LogrusLogger) Trace(args ...interface{}) {
	l.Logger.Trace(args...)
}

// Debug logs the given arguments at the debug logging level.
func (l *LogrusLogger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}

// Info logs the given arguments at the info logging level.
func (l *LogrusLogger) Info(args ...interface{}) {
	l.Logger.Info(args...)
}

// Warn logs the given arguments at the warning logging level.
func (l *LogrusLogger) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
}

// Error logs the given arguments at the error logging level.
func (l *LogrusLogger) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

// WithFields returns a message logger with multiple key-value fields.
func (l *LogrusLogger) WithFields(fields ...interface{}) logger.MessageLogger {
	f := make(map[string]interface{}, len(fields)/2)
	var key, value interface{}
	for i := 0; i+1 < len(fields); i = i + 2 {
		key = fields[i]
		value = fields[i+1]
		if s, ok := key.(string); ok {
			f[s] = value
		} else if s, ok := key.(fmt.Stringer); ok {
			f[s.String()] = value
		}
	}

	return l.Logger.WithFields(f)
}
