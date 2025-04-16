package main

import (
	"api_gateway/internal/app"
	"api_gateway/internal/config"
	"api_gateway/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
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

	application, err := app.NewApp(log, cfg)
	if err != nil {
		log.Error("failed to create app", logger.Err(err))
		os.Exit(1)
	}

	go func() {
		if err := application.Run(); err != nil {
			log.Error("application run failed", logger.Err(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	application.Stop()
	log.Info("server stopped")
}
