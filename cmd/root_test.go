package cmd

import (
	"log/slog"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want slog.Level
		ok   bool
	}{
		{"empty", "", slog.LevelInfo, true},
		{"debug", "debug", slog.LevelDebug, true},
		{"debug uppercase", "DEBUG", slog.LevelDebug, true},
		{"info with spaces", " info ", slog.LevelInfo, true},
		{"warn", "warn", slog.LevelWarn, true},
		{"warning", "warning", slog.LevelWarn, true},
		{"error", "error", slog.LevelError, true},
		{"invalid", "verbose", slog.LevelInfo, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := parseLogLevel(c.in)
			if got != c.want || ok != c.ok {
				t.Fatalf("parseLogLevel(%q) = (%v, %v), want (%v, %v)", c.in, got, ok, c.want, c.ok)
			}
		})
	}
}
