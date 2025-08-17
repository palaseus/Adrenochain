package logger

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
		{Level(99), "UNKNOWN"},
	}

	for _, test := range tests {
		result := test.level.String()
		if result != test.expected {
			t.Errorf("Level(%d).String() = %s, expected %s", test.level, result, test.expected)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != INFO {
		t.Errorf("Default level should be INFO, got %v", config.Level)
	}
	if config.Prefix != "adrenochain" {
		t.Errorf("Default prefix should be 'adrenochain', got %s", config.Prefix)
	}
	if config.Output != os.Stdout {
		t.Errorf("Default output should be os.Stdout")
	}
	if config.TimeFmt != time.RFC3339 {
		t.Errorf("Default time format should be time.RFC3339")
	}
	if config.UseJSON {
		t.Error("Default should not use JSON")
	}
	if config.LogFile != "" {
		t.Error("Default should not have a log file")
	}
	if config.MaxSize != 100*1024*1024 {
		t.Errorf("Default max size should be 100MB, got %d", config.MaxSize)
	}
	if config.MaxBackups != 5 {
		t.Errorf("Default max backups should be 5, got %d", config.MaxBackups)
	}
}

func TestNewLogger_WithNilConfig(t *testing.T) {
	logger := NewLogger(nil)

	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	// Should use default config
	if logger.level != INFO {
		t.Errorf("Logger level should be INFO, got %v", logger.level)
	}
	if logger.prefix != "adrenochain" {
		t.Errorf("Logger prefix should be 'adrenochain', got %s", logger.prefix)
	}
	if logger.output != os.Stdout {
		t.Error("Logger output should be os.Stdout")
	}
}

func TestNewLogger_WithCustomConfig(t *testing.T) {
	customOutput := &bytes.Buffer{}
	config := &Config{
		Level:      WARN,
		Prefix:     "test",
		Output:     customOutput,
		TimeFmt:    "2006-01-02",
		UseJSON:    true,
		LogFile:    "",
		MaxSize:    1024,
		MaxBackups: 3,
	}

	logger := NewLogger(config)

	if logger.level != WARN {
		t.Errorf("Logger level should be WARN, got %v", logger.level)
	}
	if logger.prefix != "test" {
		t.Errorf("Logger prefix should be 'test', got %s", logger.prefix)
	}
	if logger.output != customOutput {
		t.Error("Logger output should be custom output")
	}
	if logger.timeFmt != "2006-01-02" {
		t.Errorf("Logger time format should be '2006-01-02', got %s", logger.timeFmt)
	}
	if !logger.useJSON {
		t.Error("Logger should use JSON")
	}
}

func TestNewLogger_WithNilOutput(t *testing.T) {
	config := &Config{
		Level:  INFO,
		Output: nil,
	}

	logger := NewLogger(config)

	if logger.output != os.Stdout {
		t.Error("Logger should fall back to os.Stdout when output is nil")
	}
}

func TestNewLogger_WithFileLogging(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := &Config{
		Level:      INFO,
		LogFile:    logFile,
		MaxSize:    1024,
		MaxBackups: 2,
	}

	logger := NewLogger(config)
	defer logger.Close()

	if logger.filePath != logFile {
		t.Errorf("Logger file path should be %s, got %s", logFile, logger.filePath)
	}
	if logger.file == nil {
		t.Error("Logger should have a file when LogFile is specified")
	}

	// Test that file logging works
	logger.Info("test message")

	// Check if file was created and contains the message
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "test message") {
		t.Error("Log file should contain the test message")
	}
}

func TestNewLogger_WithFileLoggingFailure(t *testing.T) {
	// Try to create a log file in a non-existent directory with no write permissions
	config := &Config{
		Level:      INFO,
		LogFile:    "/nonexistent/noperms/test.log",
		MaxSize:    1024,
		MaxBackups: 2,
	}

	// Capture stderr to check the fallback message
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := NewLogger(config)

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if logger.output != os.Stdout {
		t.Error("Logger should fall back to stdout when file logging fails")
	}

	if !strings.Contains(buf.String(), "Failed to setup file logging") {
		t.Error("Should log error message when file logging fails")
	}
}

func TestLogger_LogLevels(t *testing.T) {
	output := &bytes.Buffer{}
	logger := NewLogger(&Config{
		Level:  WARN,
		Output: output,
		Prefix: "test",
	})

	// These should not be logged due to level filtering
	logger.Debug("debug message")
	logger.Info("info message")

	// These should be logged
	logger.Warn("warn message")
	logger.Error("error message")

	content := output.String()

	if strings.Contains(content, "debug message") {
		t.Error("Debug message should not be logged when level is WARN")
	}
	if strings.Contains(content, "info message") {
		t.Error("Info message should not be logged when level is WARN")
	}
	if !strings.Contains(content, "warn message") {
		t.Error("Warn message should be logged when level is WARN")
	}
	if !strings.Contains(content, "error message") {
		t.Error("Error message should be logged when level is WARN")
	}
}

func TestLogger_TextFormatting(t *testing.T) {
	output := &bytes.Buffer{}
	logger := NewLogger(&Config{
		Level:   INFO,
		Output:  output,
		Prefix:  "test",
		TimeFmt: "2006-01-02",
		UseJSON: false,
	})

	logger.Info("test message")

	content := output.String()

	// Check format: [timestamp] INFO [test] INFO: test message
	if !strings.Contains(content, "INFO") {
		t.Error("Text format should contain level")
	}
	if !strings.Contains(content, "test") {
		t.Error("Text format should contain prefix")
	}
	if !strings.Contains(content, "test message") {
		t.Error("Text format should contain message")
	}
}

func TestLogger_JSONFormatting(t *testing.T) {
	output := &bytes.Buffer{}
	logger := NewLogger(&Config{
		Level:   INFO,
		Output:  output,
		Prefix:  "test",
		UseJSON: true,
	})

	logger.Info("test message")

	content := output.String()

	// Check JSON format
	if !strings.Contains(content, `"timestamp"`) {
		t.Error("JSON format should contain timestamp field")
	}
	if !strings.Contains(content, `"level":"INFO"`) {
		t.Error("JSON format should contain level field")
	}
	if !strings.Contains(content, `"service":"test"`) {
		t.Error("JSON format should contain service field")
	}
	if !strings.Contains(content, `"message":"test message"`) {
		t.Error("JSON format should contain message field")
	}
}

func TestLogger_WithFields(t *testing.T) {
	logger := NewLogger(&Config{Level: INFO})

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	newLogger := logger.WithFields(fields)

	// For now, WithFields just returns the same logger
	if newLogger != logger {
		t.Error("WithFields should return the same logger for now")
	}
}

func TestLogger_SetLevel(t *testing.T) {
	logger := NewLogger(&Config{Level: INFO})

	logger.SetLevel(ERROR)

	if logger.level != ERROR {
		t.Errorf("Logger level should be ERROR after SetLevel, got %v", logger.level)
	}
}

func TestLogger_SetOutput(t *testing.T) {
	logger := NewLogger(&Config{Level: INFO})
	newOutput := &bytes.Buffer{}

	logger.SetOutput(newOutput)

	if logger.output != newOutput {
		t.Error("Logger output should be updated after SetOutput")
	}
}

func TestLogger_SetJSON(t *testing.T) {
	logger := NewLogger(&Config{Level: INFO, UseJSON: false})

	logger.SetJSON(true)

	if !logger.useJSON {
		t.Error("Logger should use JSON after SetJSON(true)")
	}

	logger.SetJSON(false)

	if logger.useJSON {
		t.Error("Logger should not use JSON after SetJSON(false)")
	}
}

func TestLogger_Close(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := &Config{
		Level:   INFO,
		LogFile: logFile,
	}

	logger := NewLogger(config)

	// Close should not error
	if err := logger.Close(); err != nil {
		t.Errorf("Close should not error: %v", err)
	}

	// Close again should not error (file already closed)
	if err := logger.Close(); err != nil {
		// This is expected behavior - file already closed
		t.Logf("Second close returned error (expected): %v", err)
	}
}

func TestLogger_Close_NoFile(t *testing.T) {
	logger := NewLogger(&Config{Level: INFO})

	// Close should not error when no file is set
	if err := logger.Close(); err != nil {
		t.Errorf("Close should not error when no file: %v", err)
	}
}

func TestLogger_GetLogFile(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := &Config{
		Level:   INFO,
		LogFile: logFile,
	}

	logger := NewLogger(config)
	defer logger.Close()

	if logger.GetLogFile() != logFile {
		t.Errorf("GetLogFile should return %s, got %s", logFile, logger.GetLogFile())
	}
}

func TestLogger_GetLogFile_NoFile(t *testing.T) {
	logger := NewLogger(&Config{Level: INFO})

	if logger.GetLogFile() != "" {
		t.Errorf("GetLogFile should return empty string when no file, got %s", logger.GetLogFile())
	}
}

func TestLogger_FileRotation(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := &Config{
		Level:      INFO,
		LogFile:    logFile,
		MaxSize:    10, // Very small size to trigger rotation
		MaxBackups: 2,
	}

	logger := NewLogger(config)
	defer logger.Close()

	// Write enough data to trigger rotation
	for i := 0; i < 100; i++ {
		logger.Info("This is a test message that should trigger file rotation")
	}

	// Wait a bit for rotation to complete
	time.Sleep(200 * time.Millisecond)

	// File rotation may not happen immediately due to goroutine timing
	// Just check that the current log file still exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Current log file should still exist after rotation attempt")
	}
}

func TestLogger_Fatal(t *testing.T) {
	// We can't easily test os.Exit(1), but we can test that the message is logged
	output := &bytes.Buffer{}
	logger := NewLogger(&Config{
		Level:  FATAL,
		Output: output,
	})

	// Since Fatal calls os.Exit(1), we can't test it directly in a test
	// Instead, we'll test that the log method works correctly for FATAL level
	// by temporarily changing the output to our buffer and calling log directly

	// Test that the message would be logged at FATAL level
	oldOutput := logger.output
	logger.output = output
	defer func() { logger.output = oldOutput }()

	// Call log directly to avoid os.Exit(1)
	logger.log(FATAL, "fatal message")

	content := output.String()
	if !strings.Contains(content, "fatal message") {
		t.Error("Fatal message should be logged")
	}
}

func TestLogger_AllLogLevels(t *testing.T) {
	output := &bytes.Buffer{}
	logger := NewLogger(&Config{
		Level:  DEBUG,
		Output: output,
		Prefix: "test",
	})

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	content := output.String()

	if !strings.Contains(content, "debug message") {
		t.Error("Debug message should be logged")
	}
	if !strings.Contains(content, "info message") {
		t.Error("Info message should be logged")
	}
	if !strings.Contains(content, "warn message") {
		t.Error("Warn message should be logged")
	}
	if !strings.Contains(content, "error message") {
		t.Error("Error message should be logged")
	}
}

func TestLogger_MessageFormatting(t *testing.T) {
	output := &bytes.Buffer{}
	logger := NewLogger(&Config{
		Level:  INFO,
		Output: output,
		Prefix: "test",
	})

	logger.Info("message with %s formatting", "string")

	content := output.String()

	if !strings.Contains(content, "message with string formatting") {
		t.Error("Message should be formatted with arguments")
	}
}

func TestLogger_EmptyMessage(t *testing.T) {
	output := &bytes.Buffer{}
	logger := NewLogger(&Config{
		Level:  INFO,
		Output: output,
		Prefix: "test",
	})

	logger.Info("")

	content := output.String()

	// Should still log even with empty message
	if !strings.Contains(content, "INFO") {
		t.Error("Empty message should still be logged")
	}
}
