// Package logger provides logging functionality for ONVIF services.
package logger

import (
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)
}

// Log levels
const (
	PANIC Level = iota
	FATAL
	ERROR
	WARN
	INFO
	DEBUG
	TRACE
)

// Level represents a log level
type Level logrus.Level

// InitLogger initializes the logger with the specified level
func InitLogger(level Level) {
	logger = logrus.New()
	logger.SetLevel(logrus.Level(level))
}

// ParseLevel converts a level string (trace, debug, info, warn, error, fatal)
// to a Level. Returns INFO and an error for unrecognised strings.
func ParseLevel(s string) (Level, error) {
	l, err := logrus.ParseLevel(s)
	return Level(l), err
}

// Fatalf logs a fatal error message
func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

// Errorf logs an error message
func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Warnf logs a warning message
func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

// Infof logs an info message
func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Debugf logs a debug message
func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

// Tracef logs a trace message
func Tracef(format string, args ...interface{}) {
	logger.Tracef(format, args...)
}
