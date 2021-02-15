package logger

type Logger interface {
	MessageLogger
	FieldLogger
}

type FieldLogger interface {
	WithFields(fields ...interface{}) MessageLogger
}

type MessageLogger interface {
	Errorf(format string, args ...interface{})
	Error(args ...interface{})
	Warnf(format string, args ...interface{})
	Warn(args ...interface{})
	Infof(format string, args ...interface{})
	Info(args ...interface{})
	Debugf(format string, args ...interface{})
	Debug(args ...interface{})
	Tracef(format string, args ...interface{})
	Trace(args ...interface{})
}
