package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	grpc2 "storage_service/internal/app/grpc"
	"storage_service/internal/config"
	"storage_service/internal/consumer/kafka"
)

type App struct {
	cfg           *config.Config
	log           *slog.Logger
	ImageConsumer *kafka.ImageConsumer
	TextConsumer  *kafka.TextConsumer
	gRPCServer    *grpc2.Server
}

func NewApp(log *slog.Logger, cfg *config.Config) (*App, error) {
	ctx := context.Background()
	s, err := grpc2.NewServer(cfg, log)
	if err != nil {
		log.Error("failed to create gRPCserver")
		return nil, err
	}
	img, err := kafka.NewImageConsumer(ctx, cfg, log)
	if err != nil {
		return nil, err
	}
	txt, err := kafka.NewTextConsumer(ctx, cfg, log)
	if err != nil {
		return nil, err
	}
	return &App{
		ImageConsumer: img,
		TextConsumer:  txt,
		cfg:           cfg,
		log:           log,
		gRPCServer:    s,
	}, nil
}

func (a *App) Run() {
	const op = "grpcapp.Run"

	a.log.Info("starting app")
	a.log.Info(fmt.Sprintf(":%d", a.cfg.GRPC.Port))
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.GRPC.Port))
	if err != nil {
		a.log.Error("net listen", err.Error())
		a.log.Error(fmt.Errorf("%s: %w", op, err).Error())
	}

	a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))
	go func() {
		if err := a.gRPCServer.Start(l); err != nil {
			a.log.Error("server grpc", err.Error())
			a.log.Error(fmt.Errorf("%s: %w", op, err).Error())
		}
	}()
	ctx := context.Background()
	go func() {
		err := a.ImageConsumer.ConsumeImages(ctx)
		if err != nil {
			a.log.Error(err.Error())
		}
	}()
	go func() {
		err := a.TextConsumer.ConsumeText(ctx)
		if err != nil {
			a.log.Error(err.Error())
		}
	}()
}

func (a *App) Stop() {
	a.log.Info("stopping app")
	err := a.TextConsumer.Close()
	if err != nil {
		a.log.Error(err.Error())
	}
	err = a.ImageConsumer.Close()
	if err != nil {
		a.log.Error(err.Error())
	}
}
