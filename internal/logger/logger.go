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

