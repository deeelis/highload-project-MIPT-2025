package grpc_server

import (
	"auth_service/internal/config"
	controller "auth_service/internal/controllers/grpc"
	"auth_service/logger"
	auth "github.com/deeelis/auth-protos/gen/go/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"net"
	"time"
)

type Server struct {
	cfg    *config.Config
	log    *slog.Logger
	Server *grpc.Server
}

func NewServer(cfg *config.Config, log *slog.Logger) (*Server, error) {
	const op = "grpc_server.NewServer"
	log = log.With(slog.String("op", op))
	log.Info("initializing gRPC server",
		slog.String("environment", cfg.Env),
		slog.Int("port", cfg.GRPC.Port))

	startTime := time.Now()
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor()))

	authController, err := controller.NewAuthController(cfg, log)
	if err != nil {
		log.Error("failed to create auth controller",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return nil, err
	}

	auth.RegisterAuthServiceServer(s, authController)
	log.Debug("auth service registered")

	if cfg.Env == "local" || cfg.Env == "dev" {
		reflection.Register(s)
		log.Debug("gRPC reflection enabled")
	}

	log.Info("gRPC server initialized successfully",
		slog.Duration("duration", time.Since(startTime)))

	return &Server{Server: s, log: log}, nil
}

func (s *Server) Start(lis net.Listener) error {
	const op = "grpc_server.Start"
	log := s.log.With(slog.String("op", op))

	log.Info("starting gRPC server",
		slog.String("address", lis.Addr().String()))

	startTime := time.Now()
	err := s.Server.Serve(lis)
	if err != nil {
		log.Error("gRPC server failed",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return err
	}

	return nil
}

func (s *Server) Stop() {
	const op = "grpc_server.Stop"
	log := s.log.With(slog.String("op", op))

	log.Info("initiating graceful shutdown")

	startTime := time.Now()
	s.Server.GracefulStop()

	log.Info("gRPC server stopped gracefully",
		slog.Duration("duration", time.Since(startTime)))
}
