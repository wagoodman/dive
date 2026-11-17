package log

import (
	"io"

	"github.com/anchore/go-logger"
	"github.com/anchore/go-logger/adapter/discard"
)

// log is the singleton used to facilitate logging internally within
var log = discard.New()

// Set replaces the default logger with the provided logger.
func Set(l logger.Logger) {
	log = l
}

// Get returns the current logger instance.
func Get() logger.Logger {
	return log
}

// file is the configured log file (nil without one), which the logger writes to alongside the console
var file io.Writer

// SetFile records the log file the logger tees to, so a UI can mute the console half and keep the file.
func SetFile(w io.Writer) {
	file = w
}

// File returns the log file set with SetFile, nil when logs only go to the console.
func File() io.Writer {
	return file
}

// Errorf takes a formatted template string and template arguments for the error logging level.
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Error logs the given arguments at the error logging level.
func Error(args ...interface{}) {
	log.Error(args...)
}

// Warnf takes a formatted template string and template arguments for the warning logging level.
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Warn logs the given arguments at the warning logging level.
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Infof takes a formatted template string and template arguments for the info logging level.
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Info logs the given arguments at the info logging level.
func Info(args ...interface{}) {
	log.Info(args...)
}

// Debugf takes a formatted template string and template arguments for the debug logging level.
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Debug logs the given arguments at the debug logging level.
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Tracef takes a formatted template string and template arguments for the trace logging level.
func Tracef(format string, args ...interface{}) {
	log.Tracef(format, args...)
}

// Trace logs the given arguments at the trace logging level.
func Trace(args ...interface{}) {
	log.Trace(args...)
}

// WithFields returns a message logger with multiple key-value fields.
func WithFields(fields ...interface{}) logger.MessageLogger {
	return log.WithFields(fields...)
}

// Nested returns a new logger with hard coded key-value pairs
func Nested(fields ...interface{}) logger.Logger {
	return log.Nested(fields...)
}
