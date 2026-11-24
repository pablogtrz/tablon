package config

import (
	"log/slog"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Setenv("TABLON_HOST", "127.0.0.1")
	t.Setenv("TABLON_PORT", "9090")
	t.Setenv("TABLON_DATA_DIR", "/tmp/tablon-test")
	t.Setenv("TABLON_API_TOKEN", "secret")
	t.Setenv("TABLON_LOG_LEVEL", "debug")
	t.Setenv("TABLON_IDLE_ENABLED", "true")
	t.Setenv("TABLON_IDLE_TIMEOUT", "90s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Host != "127.0.0.1" || cfg.Port != 9090 {
		t.Fatalf("unexpected address: %s:%d", cfg.Host, cfg.Port)
	}
	if cfg.APIToken != "secret" {
		t.Fatalf("token = %q", cfg.APIToken)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Fatalf("log level = %v", cfg.LogLevel)
	}
	if !cfg.IdleEnabled {
		t.Fatal("idle should be enabled")
	}
	if cfg.IdleTimeout != 90*time.Second {
		t.Fatalf("idle timeout = %v", cfg.IdleTimeout)
	}
	if !cfg.IsLoopback() {
		t.Fatal("127.0.0.1 should be loopback")
	}
	if got, want := cfg.DBPath(), filepath.Join("/tmp/tablon-test", "app.db"); got != want {
		t.Fatalf("db path = %q, want %q", got, want)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "port", key: "TABLON_PORT", value: "not-a-port"},
		{name: "port range", key: "TABLON_PORT", value: "0"},
		{name: "log level", key: "TABLON_LOG_LEVEL", value: "trace"},
		{name: "data dir", key: "TABLON_DATA_DIR", value: " "},
		{name: "idle enabled", key: "TABLON_IDLE_ENABLED", value: "sometimes"},
		{name: "idle timeout", key: "TABLON_IDLE_TIMEOUT", value: "soon"},
		{name: "idle timeout zero", key: "TABLON_IDLE_TIMEOUT", value: "0s"},
		{name: "idle timeout negative", key: "TABLON_IDLE_TIMEOUT", value: "-1s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TABLON_HOST", "127.0.0.1")
			t.Setenv("TABLON_PORT", "8448")
			t.Setenv("TABLON_DATA_DIR", "./data")
			t.Setenv("TABLON_LOG_LEVEL", "info")
			t.Setenv("TABLON_IDLE_ENABLED", "true")
			t.Setenv("TABLON_IDLE_TIMEOUT", "5m")
			t.Setenv(tt.key, tt.value)
			if _, err := Load(); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestIsLoopback(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"127.0.0.1", true},
		{"localhost", true},
		{"::1", true},
		{"0.0.0.0", false},
		{"192.168.1.10", false},
	}
	for _, tt := range tests {
		cfg := Config{Host: tt.host}
		if got := cfg.IsLoopback(); got != tt.want {
			t.Errorf("IsLoopback(%q) = %v, want %v", tt.host, got, tt.want)
		}
	}
}
