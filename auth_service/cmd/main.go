package main

import (
	app2 "auth_service/internal/app"
	"auth_service/internal/config"
	"auth_service/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	startTime := time.Now()
	cfg, err := config.MustLoad()
	if err != nil {
		slog.Error("failed to load configuration", logger.Err(err))
		os.Exit(1)
	}

	log := logger.SetUpLogger(cfg.Env)
	log.Info("starting authentication service",
		slog.String("environment", cfg.Env),
		slog.Int("grpc_port", cfg.GRPC.Port),
	)

	app, err := app2.NewApp(log, cfg)
	if err != nil {
		log.Error("application initialization failed",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		os.Exit(1)
	}
	log.Info("application initialized successfully",
		slog.Duration("init_duration", time.Since(startTime)))

	go func() {
		log.Info("starting application services")
		if err := app.Run(); err != nil {
			log.Error("application runtime error",
				logger.Err(err),
				slog.Duration("uptime", time.Since(startTime)))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-quit

	log.Info("shutdown signal received",
		slog.String("signal", sig.String()),
		slog.Duration("uptime", time.Since(startTime)))

	log.Info("initiating graceful shutdown...")
	app.Stop()
	log.Info("application stopped gracefully",
		slog.Duration("total_uptime", time.Since(startTime)))
}
