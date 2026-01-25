package config

import (
	"log/slog"
	"testing"
)

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		expected slog.Level
	}{
		{"debug level", "debug", slog.LevelDebug},
		{"info level", "info", slog.LevelInfo},
		{"warn level", "warn", slog.LevelWarn},
		{"warning level", "warning", slog.LevelWarn},
		{"error level", "error", slog.LevelError},
		{"uppercase", "DEBUG", slog.LevelDebug},
		{"mixed case", "Info", slog.LevelInfo},
		{"invalid defaults to info", "invalid", slog.LevelInfo},
		{"empty defaults to info", "", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{LOG_LEVEL: tt.logLevel}
			got := cfg.GetLogLevel()
			if got != tt.expected {
				t.Errorf("GetLogLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsJSONFormat(t *testing.T) {
	tests := []struct {
		name      string
		logFormat string
		expected  bool
	}{
		{"json format", "json", true},
		{"uppercase JSON", "JSON", true},
		{"mixed case", "Json", true},
		{"text format", "text", false},
		{"empty", "", false},
		{"invalid", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{LOG_FORMAT: tt.logFormat}
			got := cfg.IsJSONFormat()
			if got != tt.expected {
				t.Errorf("IsJSONFormat() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSetupLogger(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		logFormat string
	}{
		{"json debug", "debug", "json"},
		{"text info", "info", "text"},
		{"json error", "error", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				LOG_LEVEL:  tt.logLevel,
				LOG_FORMAT: tt.logFormat,
			}
			logger := cfg.SetupLogger()
			if logger == nil {
				t.Error("SetupLogger() returned nil")
			}
		})
	}
}

func TestGetRedisAddress(t *testing.T) {
	cfg := &Config{
		REDIS_HOST: "localhost",
		REDIS_PORT: "6379",
	}
	expected := "localhost:6379"
	got := cfg.GetRedisAddress()
	if got != expected {
		t.Errorf("GetRedisAddress() = %v, want %v", got, expected)
	}
}

func TestGetRedisURL(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     string
		password string
		db       int
		expected string
	}{
		{
			"no password",
			"localhost",
			"6379",
			"",
			0,
			"redis://localhost:6379/0",
		},
		{
			"with password",
			"localhost",
			"6379",
			"secret",
			1,
			"redis://:secret@localhost:6379/1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				REDIS_HOST:     tt.host,
				REDIS_PORT:     tt.port,
				REDIS_PASSWORD: tt.password,
				REDIS_DB:       tt.db,
			}
			got := cfg.GetRedisURL()
			if got != tt.expected {
				t.Errorf("GetRedisURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}
