package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents the level of logging
type LogLevel int

const (
	FATAL LogLevel = iota
	ERROR
	WARN
	INFO
	DEBUG
	TRACE
)

var (
	currentLevel LogLevel
	logger       *log.Logger
)

// InitLogger initializes the logger with the specified level
func InitLogger(level LogLevel) {
	currentLevel = level
	logger = log.New(os.Stderr, "", log.LstdFlags)
}

// SetLevel sets the logging level
func SetLevel(level LogLevel) {
	currentLevel = level
}

// Fatal logs a fatal error message
func Fatal(format string, v ...interface{}) {
	if currentLevel >= FATAL {
		logger.Printf("[FATAL] "+format, v...)
		os.Exit(1)
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if currentLevel >= ERROR {
		logger.Printf("[ERROR] "+format, v...)
	}
}

// Warn logs a warning message
func Warn(format string, v ...interface{}) {
	if currentLevel >= WARN {
		logger.Printf("[WARN] "+format, v...)
	}
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	if currentLevel >= INFO {
		logger.Printf("[INFO] "+format, v...)
	}
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	if currentLevel >= DEBUG {
		logger.Printf("[DEBUG] "+format, v...)
	}
}

// Trace logs a trace message
func Trace(format string, v ...interface{}) {
	if currentLevel >= TRACE {
		logger.Printf("[TRACE] "+format, v...)
	}
}

// LogWithTime logs a message with a timestamp
func LogWithTime(level LogLevel, format string, v ...interface{}) {
	if currentLevel >= level {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		message := fmt.Sprintf(format, v...)
		logger.Printf("%s %s", timestamp, message)
	}
}
