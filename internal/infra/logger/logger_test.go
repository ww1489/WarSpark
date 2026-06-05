package logger

import (
	"log/slog"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		value string
		level slog.Level
	}{
		{name: "debug", value: "debug", level: slog.LevelDebug},
		{name: "info", value: "info", level: slog.LevelInfo},
		{name: "empty", value: "", level: slog.LevelInfo},
		{name: "warn", value: "warn", level: slog.LevelWarn},
		{name: "warning", value: "warning", level: slog.LevelWarn},
		{name: "error", value: "error", level: slog.LevelError},
		{name: "trim and case", value: " ERROR ", level: slog.LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, err := parseLogLevel(tt.value)
			if err != nil {
				t.Fatalf("parseLogLevel returned error: %v", err)
			}
			if level != tt.level {
				t.Fatalf("expected level %s, got %s", tt.level, level)
			}
		})
	}
}

func TestParseLogLevelRejectsUnsupportedValue(t *testing.T) {
	_, err := parseLogLevel("verbose")
	if err == nil {
		t.Fatal("expected unsupported log level to return error")
	}
}
