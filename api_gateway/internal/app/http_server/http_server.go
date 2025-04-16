package http_server

import (
	"api_gateway/internal/controllers/router"
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"api_gateway/internal/config"
	"api_gateway/logger"
)

type HTTPServer struct {
	cfg    *config.Config
	server *http.Server
	log    *slog.Logger
}

func NewHTTPServer(log *slog.Logger, cfg *config.Config) (*HTTPServer, error) {
	router, err := http_controllers.NewRouter(cfg, log)
	if err != nil {
		return nil, err
	}

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	return &HTTPServer{
		server: server,
		log:    log,
		cfg:    cfg,
	}, nil
}

func (s *HTTPServer) Start() error {
	s.log.Info("starting HTTP server", slog.String("addr", s.server.Addr))
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.log.Error("HTTP server shutdown error", logger.Err(err))
	}
}
