package grpc_server

import (
	"auth_service/internal/config"
	controller "auth_service/internal/controllers/grpc"
	auth "github.com/deeelis/auth-protos/gen/go/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"net"
)

type Server struct {
	cfg    *config.Config
	log    *slog.Logger
	Server *grpc.Server
}

func NewServer(cfg *config.Config, log *slog.Logger) (*Server, error) {
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor()))
	authController, err := controller.NewAuthController(cfg, log)
	if err != nil {
		log.Error("creating new server error", err.Error())
		return nil, err
	}
	auth.RegisterAuthServiceServer(s, authController)
	if cfg.Env == "local" || cfg.Env == "dev" {
		reflection.Register(s)
	}
	return &Server{Server: s, log: log}, nil
}

func (s *Server) Start(lis net.Listener) error {
	s.log.Info("start server")
	return s.Server.Serve(lis)
}

func (s *Server) Stop() {
	s.log.Info("stop server")
	s.Server.GracefulStop()
}
