package main

import (
	_ "github.com/lib/pq"
	"log/slog"
	"os"
	"os/signal"
	"storage_service/internal/app"
	"storage_service/internal/config"
	"storage_service/logger"
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
	)

	application, err := app.NewApp(log, cfg)
	if err != nil {
		log.Error("failed to create app", logger.Err(err))
		os.Exit(1)
	}

	go application.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	application.Stop()
	log.Info("server stopped")

}
