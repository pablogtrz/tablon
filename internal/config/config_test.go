package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func unsetenv(t *testing.T, key string) {
	t.Helper()
	old, ok := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
	t.Cleanup(func() {
		if ok {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

func TestLoadDefaults(t *testing.T) {
	for _, key := range []string{
		"TABLON_HOST", "TABLON_PORT", "TABLON_DATA_DIR", "TABLON_API_TOKEN",
		"TABLON_LOG_LEVEL", "TABLON_IDLE_ENABLED", "TABLON_IDLE_TIMEOUT",
	} {
		unsetenv(t, key)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Host != "127.0.0.1" || cfg.Port != 8448 {
		t.Fatalf("unexpected address: %s:%d", cfg.Host, cfg.Port)
	}
	if cfg.DataDir != "./data" {
		t.Fatalf("data dir = %q", cfg.DataDir)
	}
	if cfg.APIToken != "" {
		t.Fatalf("token = %q", cfg.APIToken)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Fatalf("log level = %v", cfg.LogLevel)
	}
	if !cfg.IdleEnabled {
		t.Fatal("idle should be enabled by default")
	}
	if cfg.IdleTimeout != 5*time.Minute {
		t.Fatalf("idle timeout = %v", cfg.IdleTimeout)
	}
}

func TestLoadIdleDisabled(t *testing.T) {
	for _, key := range []string{
		"TABLON_HOST", "TABLON_PORT", "TABLON_DATA_DIR", "TABLON_API_TOKEN",
		"TABLON_LOG_LEVEL",
	} {
		unsetenv(t, key)
	}
	t.Setenv("TABLON_IDLE_ENABLED", "false")
	t.Setenv("TABLON_IDLE_TIMEOUT", "90s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.IdleEnabled {
		t.Fatal("idle should be disabled")
	}
	if cfg.IdleTimeout != 90*time.Second {
		t.Fatalf("idle timeout = %v", cfg.IdleTimeout)
	}
}

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
