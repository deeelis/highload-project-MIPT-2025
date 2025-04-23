package main

import (
	"api_gateway/internal/app"
	"api_gateway/internal/config"
	"api_gateway/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		slog.Error("failed to load config", logger.Err(err))
		os.Exit(1)
	}

	log := logger.SetUpLogger(cfg.Env)
	log.Info("starting api gateway",
		slog.String("env", cfg.Env),
		slog.Int("port", cfg.HTTP.Port),
	)

	startTime := time.Now()
	application, err := app.NewApp(log, cfg)
	if err != nil {
		log.Error("application initialization failed",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		os.Exit(1)
	}
	log.Info("application initialized successfully",
		slog.Duration("duration", time.Since(startTime)))

	go func() {
		log.Info("starting application server")
		if err := application.Run(); err != nil {
			log.Error("application runtime error", logger.Err(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-quit

	log.Info("shutdown signal received",
		slog.String("signal", sig.String()))

	log.Info("starting graceful shutdown...")

	shutdownStart := time.Now()
	application.Stop()

	log.Info("application stopped gracefully",
		slog.Duration("shutdown_duration", time.Since(shutdownStart)),
		slog.Duration("uptime", time.Since(startTime)))
}
