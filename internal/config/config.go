package config

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultHost    = "127.0.0.1"
	DefaultPort    = 8080
	DefaultDataDir = "./data"
)

type Config struct {
	Host     string
	Port     int
	DataDir  string
	APIToken string
	LogLevel slog.Level
}

func Load() (Config, error) {
	cfg := Config{
		Host:     env("TABLON_HOST", DefaultHost),
		DataDir:  env("TABLON_DATA_DIR", DefaultDataDir),
		LogLevel: slog.LevelInfo,
	}

	portRaw := env("TABLON_PORT", strconv.Itoa(DefaultPort))
	port, err := strconv.Atoi(portRaw)
	if err != nil {
		return Config{}, fmt.Errorf("TABLON_PORT must be a number: %w", err)
	}
	cfg.Port = port
	cfg.APIToken = os.Getenv("TABLON_API_TOKEN")

	levelRaw := strings.ToLower(env("TABLON_LOG_LEVEL", "info"))
	switch levelRaw {
	case "debug":
		cfg.LogLevel = slog.LevelDebug
	case "info":
		cfg.LogLevel = slog.LevelInfo
	case "warn", "warning":
		cfg.LogLevel = slog.LevelWarn
	case "error":
		cfg.LogLevel = slog.LevelError
	default:
		return Config{}, fmt.Errorf("TABLON_LOG_LEVEL must be debug, info, warn or error")
	}

	if cfg.Host == "" {
		return Config{}, fmt.Errorf("TABLON_HOST cannot be empty")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return Config{}, fmt.Errorf("TABLON_PORT must be between 1 and 65535")
	}
	if strings.TrimSpace(cfg.DataDir) == "" {
		return Config{}, fmt.Errorf("TABLON_DATA_DIR cannot be empty")
	}
	return cfg, nil
}

func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

func (c Config) DBPath() string {
	return filepath.Join(c.DataDir, "app.db")
}

func (c Config) IsLoopback() bool {
	if strings.EqualFold(c.Host, "localhost") {
		return true
	}
	if ip := net.ParseIP(c.Host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
