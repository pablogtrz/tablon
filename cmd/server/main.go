package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tablon "github.com/pablogtrz/tablon"
	"github.com/pablogtrz/tablon/internal/api"
	"github.com/pablogtrz/tablon/internal/config"
	"github.com/pablogtrz/tablon/internal/events"
	"github.com/pablogtrz/tablon/internal/idle"
	"github.com/pablogtrz/tablon/internal/service"
	"github.com/pablogtrz/tablon/internal/store"
	"github.com/pablogtrz/tablon/internal/version"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-version") {
		fmt.Printf("tablond %s (commit %s, built %s)\n", version.Version, version.Commit, version.Date)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "configuration error:", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	st, err := store.Open(cfg.DataDir)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer st.Close()

	hub := events.NewHub(logger)
	defer hub.Close()
	svc := service.New(st, hub)
	idleCtl := idle.New(cfg.IdleEnabled, cfg.IdleTimeout, logger, func() {
		if err := st.EnterIdle(context.Background()); err != nil {
			logger.Warn("release idle resources", "error", err)
		}
	}, st.ExitIdle)
	handler := api.New(cfg, svc, hub, idleCtl, logger, tablon.Frontend, tablon.OpenAPI)

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	// Shutdown waits for active handlers. Closing the hub ends SSE streams so
	// they cannot hold the process for the full shutdown timeout.
	srv.RegisterOnShutdown(hub.Close)

	if !cfg.IsLoopback() && cfg.APIToken == "" {
		logger.Info("api is bound to a non-loopback address without TABLON_API_TOKEN")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.IdleEnabled {
		logger.Info("idle mode enabled", "timeout", cfg.IdleTimeout)
	}
	go idleCtl.Run(ctx)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("tablon listening", "addr", cfg.Addr(), "database", st.Path(), "version", version.Version)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutting down")
	case err := <-errCh:
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
