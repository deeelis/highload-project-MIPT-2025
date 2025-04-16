package main

import (
	app2 "auth_service/internal/app"
	"auth_service/internal/config"
	"auth_service/logger"
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
	log.Info("starting auth service",
		slog.String("env", cfg.Env),
		slog.Int("port", cfg.GRPC.Port),
	)

	app, err := app2.NewApp(log, cfg)
	if err != nil {
		log.Error("failed to make app", logger.Err(err))
		os.Exit(1)
	}

	err = app.Run()
	if err != nil {
		log.Error("app run", err.Error())
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	app.Stop()
	log.Info("server stopped")
}
