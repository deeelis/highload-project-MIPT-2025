package app

import (
	grpc2 "auth_service/internal/app/grpc_server"
	"auth_service/internal/config"
	"fmt"
	"log/slog"
	"net"
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
	s, err := grpc2.NewServer(cfg, log)
	if err != nil {
		log.Error("failed to create gRPCserver")
		return nil, err
	}
	return &App{
		cfg:        cfg,
		log:        log,
		gRPCServer: s,
	}, nil
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	a.log.Info(fmt.Sprintf(":%d", a.cfg.GRPC.Port))
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.GRPC.Port))
	if err != nil {
		a.log.Error("net listen", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))
	if err := a.gRPCServer.Start(l); err != nil {
		a.log.Error("server grpc", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.cfg.GRPC.Port))

	a.gRPCServer.Stop()
}
