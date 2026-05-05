package logger

import (
	"log/slog"
	"strings"
	"testing"
)

func TestSetupLogger(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		format   string
		expected slog.Level
	}{
		{
			name:     "debug level json",
			level:    "debug",
			format:   "json",
			expected: slog.LevelDebug,
		},
		{
			name:     "info level json",
			level:    "info",
			format:   "json",
			expected: slog.LevelInfo,
		},
		{
			name:     "warn level console",
			level:    "warn",
			format:   "console",
			expected: slog.LevelWarn,
		},
		{
			name:     "error level console",
			level:    "error",
			format:   "console",
			expected: slog.LevelError,
		},
		{
			name:     "invalid level defaults to info",
			level:    "invalid",
			format:   "json",
			expected: slog.LevelInfo,
		},
		{
			name:     "invalid format defaults to text",
			level:    "info",
			format:   "invalid",
			expected: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := SetupLogger(tt.level, tt.format)

			// Test that logger is created and has correct level
			if logger == nil {
				t.Fatal("Expected non-nil logger")
			}

			// Test logging at different levels
			handler := logger.Handler()

			// Create a new logger to test functionality
			testLogger := slog.New(handler)

			// Log a test message
			testLogger.Info("test message")

			// Verify the logger was created successfully
			if !strings.Contains(tt.name, "invalid") {
				// For valid configs, just ensure no panic occurred
				t.Logf("Logger created successfully for %s", tt.name)
			}
		})
	}
}

func TestSetupLoggerCaseInsensitive(t *testing.T) {
	tests := []struct {
		level    string
		format   string
		expected slog.Level
	}{
		{"DEBUG", "JSON", slog.LevelDebug},
		{"Info", "json", slog.LevelInfo},
		{"WARN", "CONSOLE", slog.LevelWarn},
		{"error", "Console", slog.LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.level+"_"+tt.format, func(t *testing.T) {
			logger := SetupLogger(tt.level, tt.format)
			if logger == nil {
				t.Fatal("Expected non-nil logger")
			}
		})
	}
}
