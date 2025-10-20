package router

import (
	"fmt"
	"log"
	"os"
)

// Logger provides a simple logging interface for the router package
type Logger struct {
	prefix   string
	logger   *log.Logger
	testMode bool
}

// NewLogger creates a new logger instance
func NewLogger(prefix string) *Logger {
	return &Logger{
		prefix:   prefix,
		logger:   log.New(os.Stdout, fmt.Sprintf("[%s] ", prefix), log.LstdFlags),
		testMode: false,
	}
}

// NewTestLogger creates a new logger instance for testing that suppresses Info logs
func NewTestLogger(prefix string) *Logger {
	return &Logger{
		prefix:   prefix,
		logger:   log.New(os.Stdout, fmt.Sprintf("[%s] ", prefix), log.LstdFlags),
		testMode: true,
	}
}

// SetTestMode sets the test mode for the logger
func (l *Logger) SetTestMode(testMode bool) {
	l.testMode = testMode
}

// Info logs an info message (suppressed in test mode)
func (l *Logger) Info(format string, args ...interface{}) {
	if !l.testMode {
		l.logger.Printf("INFO: "+format, args...)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.logger.Printf("DEBUG: "+format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.logger.Printf("WARN: "+format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.logger.Printf("ERROR: "+format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.logger.Fatalf("FATAL: "+format, args...)
}
