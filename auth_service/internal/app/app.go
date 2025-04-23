package app

import (
	grpc2 "auth_service/internal/app/grpc_server"
	"auth_service/internal/config"
	"auth_service/logger"
	"fmt"
	"log/slog"
	"net"
	"time"
)

type App struct {
	cfg        *config.Config
	log        *slog.Logger
	gRPCServer *grpc2.Server
}

func NewApp(
	log *slog.Logger,
	cfg *config.Config,
) (*App, error) {
	const op = "app.NewApp"
	log = log.With(slog.String("op", op))
	log.Info("initializing application",
		slog.String("env", cfg.Env),
		slog.Int("grpc_port", cfg.GRPC.Port))

	startTime := time.Now()

	s, err := grpc2.NewServer(cfg, log)
	if err != nil {
		log.Error("failed to initialize gRPC server",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("application initialized successfully",
		slog.Duration("duration", time.Since(startTime)))

	return &App{
		cfg:        cfg,
		log:        log,
		gRPCServer: s,
	}, nil
}

func (a *App) Run() error {
	const op = "app.Run"
	log := a.log.With(slog.String("op", op))

	log.Info("starting application services")
	startTime := time.Now()

	addr := fmt.Sprintf(":%d", a.cfg.GRPC.Port)
	log.Debug("creating TCP listener", slog.String("address", addr))

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("failed to create listener",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC server starting",
		slog.String("address", l.Addr().String()),
		slog.Duration("startup_time", time.Since(startTime)))

	if err := a.gRPCServer.Start(l); err != nil {
		log.Error("gRPC server failed to start",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("application services started successfully",
		slog.Duration("duration", time.Since(startTime)))

	return nil
}

func (a *App) Stop() {
	const op = "app.Stop"
	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.cfg.GRPC.Port),
	)

	log.Info("initiating application shutdown")
	startTime := time.Now()

	a.gRPCServer.Stop()

	log.Info("application stopped gracefully",
		slog.Duration("shutdown_duration", time.Since(startTime)))
}
