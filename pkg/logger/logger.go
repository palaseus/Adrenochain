package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	systemconfig "github.com/palaseus/adrenochain/pkg/config"
)

// Level represents the logging level
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a structured logger
type Logger struct {
	level    Level
	prefix   string
	output   io.Writer
	timeFmt  string
	useJSON  bool
	file     *os.File
	filePath string

	// SECURITY FIX: Log sanitization
	sanitizeLogs      bool
	sensitivePatterns []*regexp.Regexp
}

// Config holds logger configuration
type Config struct {
	Level      Level
	Prefix     string
	Output     io.Writer
	TimeFmt    string
	UseJSON    bool
	LogFile    string
	MaxSize    int64 // Maximum file size in bytes before rotation
	MaxBackups int   // Maximum number of backup files to keep

	// SECURITY FIX: Log sanitization configuration
	SanitizeLogs bool
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      INFO,
		Prefix:     "adrenochain",
		Output:     os.Stdout,
		TimeFmt:    time.RFC3339,
		UseJSON:    false,
		LogFile:    "",
		MaxSize:    100 * 1024 * 1024, // 100MB
		MaxBackups: 5,
	}
}

// NewLogger creates a new logger with the given configuration
func NewLogger(config *Config) *Logger {
	if config == nil {
		config = DefaultConfig()
	}

	logger := &Logger{
		level:    config.Level,
		prefix:   config.Prefix,
		output:   config.Output,
		timeFmt:  config.TimeFmt,
		useJSON:  config.UseJSON,
		filePath: config.LogFile,

		// SECURITY FIX: Initialize log sanitization
		sanitizeLogs: config.SanitizeLogs,
	}

	// Ensure output is always set
	if logger.output == nil {
		logger.output = os.Stdout
	}

	// Set up file logging if specified
	if config.LogFile != "" {
		if err := logger.setupFileLogging(config); err != nil {
			// Fall back to stdout if file logging fails
			fmt.Fprintf(os.Stderr, "Failed to setup file logging: %v, falling back to stdout\n", err)
			logger.output = os.Stdout
		}
	}

	return logger
}

// setupFileLogging sets up file logging with rotation
func (l *Logger) setupFileLogging(config *Config) error {
	// Ensure directory exists
	dir := filepath.Dir(config.LogFile)
	systemConfig := systemconfig.DefaultSystemConfig()
	filePerms := os.FileMode(systemConfig.Storage.FilePermissions)
	if err := os.MkdirAll(dir, filePerms); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.file = file
	l.output = file

	// Start file rotation goroutine
	go l.rotateLogFile(config)

	return nil
}

// rotateLogFile handles log file rotation based on size
func (l *Logger) rotateLogFile(config *Config) {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for range ticker.C {
		if l.file == nil {
			continue
		}

		// Check file size
		info, err := l.file.Stat()
		if err != nil {
			continue
		}

		if info.Size() >= config.MaxSize {
			l.rotateFile(config)
		}
	}
}

// rotateFile performs the actual file rotation
func (l *Logger) rotateFile(config *Config) {
	if l.file == nil {
		return
	}

	// Close current file
	l.file.Close()

	// Rotate backup files
	for i := config.MaxBackups - 1; i > 0; i-- {
		oldName := fmt.Sprintf("%s.%d", l.filePath, i)
		newName := fmt.Sprintf("%s.%d", l.filePath, i+1)

		if _, err := os.Stat(oldName); err == nil {
			os.Rename(oldName, newName)
		}
	}

	// Rename current file to .1
	backupName := fmt.Sprintf("%s.1", l.filePath)
	os.Rename(l.filePath, backupName)

	// Open new log file
	file, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fall back to stdout if rotation fails
		l.output = os.Stdout
		return
	}

	l.file = file
	l.output = file
}

// log formats and writes a log message
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format(l.timeFmt)
	message := fmt.Sprintf(format, args...)

	if l.useJSON {
		l.logJSON(level, timestamp, message)
	} else {
		l.logText(level, timestamp, message)
	}
}

// SECURITY FIX: sanitizeLogMessage removes sensitive information from log messages
func (l *Logger) sanitizeLogMessage(message string) string {
	if !l.sanitizeLogs {
		return message
	}

	// Initialize sensitive patterns if not already done
	if l.sensitivePatterns == nil {
		l.initSensitivePatterns()
	}

	sanitized := message

	// Apply all sensitive pattern replacements
	for _, pattern := range l.sensitivePatterns {
		sanitized = pattern.ReplaceAllString(sanitized, "[REDACTED]")
	}

	return sanitized
}

// SECURITY FIX: initSensitivePatterns initializes regex patterns for sensitive data
func (l *Logger) initSensitivePatterns() {
	l.sensitivePatterns = []*regexp.Regexp{
		// Private keys (hex format)
		regexp.MustCompile(`(?i)(private[_-]?key|privkey|secret[_-]?key)\s*[:=]\s*[a-f0-9]{64,}`),

		// API keys and tokens
		regexp.MustCompile(`(?i)(api[_-]?key|token|auth[_-]?token)\s*[:=]\s*[a-zA-Z0-9_-]{20,}`),

		// Passwords
		regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*[^\s]{3,}`),

		// Credit card numbers
		regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`),

		// Social Security Numbers
		regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),

		// Email addresses (partial redaction)
		regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`),

		// IP addresses (internal/private)
		regexp.MustCompile(`\b(?:10\.|172\.(?:1[6-9]|2\d|3[01])\.|192\.168\.)\d{1,3}\.\d{1,3}\b`),

		// Database connection strings
		regexp.MustCompile(`(?i)(mysql|postgres|mongodb)://[^\s]+`),

		// JWT tokens
		regexp.MustCompile(`eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*`),

		// Bitcoin addresses
		regexp.MustCompile(`\b[13][a-km-zA-HJ-NP-Z1-9]{25,34}\b`),

		// Ethereum addresses
		regexp.MustCompile(`\b0x[a-fA-F0-9]{40}\b`),
	}
}

// logText writes a text-formatted log message
func (l *Logger) logText(level Level, timestamp, message string) {
	// SECURITY FIX: Sanitize log message before output
	sanitizedMessage := l.sanitizeLogMessage(message)
	fmt.Fprintf(l.output, "[%s] %s [%s] %s: %s\n",
		timestamp, level.String(), l.prefix, level.String(), sanitizedMessage)
}

// logJSON writes a JSON-formatted log message
func (l *Logger) logJSON(level Level, timestamp, message string) {
	// SECURITY FIX: Sanitize log message before output
	sanitizedMessage := l.sanitizeLogMessage(message)
	// Simple JSON format for now
	jsonMsg := fmt.Sprintf(`{"timestamp":"%s","level":"%s","service":"%s","message":"%s"}`,
		timestamp, level.String(), l.prefix, sanitizedMessage)
	fmt.Fprintln(l.output, jsonMsg)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}

// WithFields creates a new logger with additional context fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	// For now, just return the same logger
	// In a more advanced implementation, this would add fields to JSON output
	return l
}

// SetLevel changes the logging level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// SetOutput changes the output writer
func (l *Logger) SetOutput(output io.Writer) {
	l.output = output
}

// SetJSON enables or disables JSON output
func (l *Logger) SetJSON(useJSON bool) {
	l.useJSON = useJSON
}

// Close closes the logger and any open files
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// GetLogFile returns the current log file path
func (l *Logger) GetLogFile() string {
	return l.filePath
}
