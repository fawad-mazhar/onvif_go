package logger

import (
	"fmt"
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

// SetLevel sets the logging level
func SetLevel(level Level) {
	logger.SetLevel(logrus.Level(level))
}

// Fatal logs a fatal error message
func Fatal(format string, args ...interface{}) {
	if len(args) > 0 {
		logger.Fatal(fmt.Sprintf(format, args...))
	} else {
		logger.Fatal(format)
	}
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	if len(args) > 0 {
		logger.Error(fmt.Sprintf(format, args...))
	} else {
		logger.Error(format)
	}
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	if len(args) > 0 {
		logger.Warn(fmt.Sprintf(format, args...))
	} else {
		logger.Warn(format)
	}
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	if len(args) > 0 {
		logger.Info(fmt.Sprintf(format, args...))
	} else {
		logger.Info(format)
	}
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	if len(args) > 0 {
		logger.Debug(fmt.Sprintf(format, args...))
	} else {
		logger.Debug(format)
	}
}

// Trace logs a trace message
func Trace(format string, args ...interface{}) {
	if len(args) > 0 {
		logger.Trace(fmt.Sprintf(format, args...))
	} else {
		logger.Trace(format)
	}
}

// Fatalf logs a fatal error message with formatting
func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

// Errorf logs an error message with formatting
func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Warnf logs a warning message with formatting
func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

// Infof logs an info message with formatting
func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Debugf logs a debug message with formatting
func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

// Tracef logs a trace message with formatting
func Tracef(format string, args ...interface{}) {
	logger.Tracef(format, args...)
}
