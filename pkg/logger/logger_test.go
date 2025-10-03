package logger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestLoggerInit(t *testing.T) {
	var buf bytes.Buffer
	InitWithWriter(&buf, "info")

	// Test that logger is initialized
	if zerolog.GlobalLevel() != zerolog.InfoLevel {
		t.Errorf("Expected global log level to be InfoLevel, got %v", zerolog.GlobalLevel())
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zerolog.Level
	}{
		{"trace", zerolog.TraceLevel},
		{"TRACE", zerolog.TraceLevel},
		{"debug", zerolog.DebugLevel},
		{"DEBUG", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"INFO", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"warning", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"ERROR", zerolog.ErrorLevel},
		{"invalid", zerolog.InfoLevel}, // Default
		{"", zerolog.InfoLevel},        // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoggerOutput(t *testing.T) {
	var buf bytes.Buffer
	InitWithWriter(&buf, "info")

	Logger.Info().Str("test", "value").Msg("Test message")

	output := buf.String()
	if !strings.Contains(output, "Test message") {
		t.Errorf("Expected output to contain 'Test message', got: %s", output)
	}
	if !strings.Contains(output, "test") {
		t.Errorf("Expected output to contain 'test' field, got: %s", output)
	}
}

func TestMetricsHookIntegration(t *testing.T) {
	// Test that MetricsHook can be set (for future Prometheus integration)
	var buf bytes.Buffer
	
	// Simple hook that does nothing but validates it can be set
	testHook := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
		// Hook functionality would be implemented here for metrics
	})
	
	SetMetricsHook(testHook)
	InitWithWriter(&buf, "info")
	
	// Verify logger works with hook
	Logger.Info().Msg("Test with hook")
	
	if !strings.Contains(buf.String(), "Test with hook") {
		t.Error("Logger with hook did not produce expected output")
	}
}
