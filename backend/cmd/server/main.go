package main

import (
	"context"
	"github.com/frankyangcl/ai-support-agent/backend/internal/config"
	"github.com/frankyangcl/ai-support-agent/backend/internal/database"
	"github.com/frankyangcl/ai-support-agent/backend/internal/router"
	appserver "github.com/frankyangcl/ai-support-agent/backend/internal/server"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, err := database.Connect(ctx, cfg.DatabaseURL, cfg.DBMaxOpen, cfg.DBMaxIdle, cfg.DBConnMaxLifetime)
	if err != nil {
		logger.Error("database startup failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	engine, err := router.Setup(db, cfg)
	if err != nil {
		logger.Error("router startup failed", "error", err)
		os.Exit(1)
	}
	srv := &http.Server{Addr: cfg.ServerAddr, Handler: engine, ReadHeaderTimeout: cfg.ReadHeaderTimeout, IdleTimeout: cfg.IdleTimeout, WriteTimeout: 0}
	logger.Info("server_started", "address", cfg.ServerAddr, "gin_mode", cfg.GinMode)
	if err = appserver.Serve(ctx, srv, cfg.ShutdownTimeout); err != nil {
		logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
	logger.Info("server_stopped")
}
