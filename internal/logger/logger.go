// Package logger provides logging functionality for ONVIF services.
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

// Fatalf logs a fatal error message
func Fatalf(format string, args ...interface{}) {
	if len(args) > 0 {
		logger.Fatalf(fmt.Sprintf(format, args...))
	} else {
		logger.Fatalf(format)
	}
}

// Errorf logs an error message
func Errorf(format string, args ...interface{}) {
	if len(args) > 0 {
		logger.Errorf(fmt.Sprintf(format, args...))
	} else {
		logger.Errorf(format)
	}
}

// Warnf logs a warning message
func Warnf(format string, args ...interface{}) {
	if len(args) > 0 {
		logger.Warnf(fmt.Sprintf(format, args...))
	} else {
		logger.Warnf(format)
	}
}

// Infof logs an info message
func Infof(format string, args ...interface{}) {
	if len(args) > 0 {
		logger.Infof(fmt.Sprintf(format, args...))
	} else {
		logger.Infof(format)
	}
}

// Debugf logs a debug message
func Debugf(format string, args ...interface{}) {
	if len(args) > 0 {
		logger.Debugf(fmt.Sprintf(format, args...))
	} else {
		logger.Debugf(format)
	}
}

