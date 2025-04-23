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
	const op = "http_server.NewHTTPServer"
	log = log.With(slog.String("op", op))

	log.Info("creating new HTTP server instance",
		slog.Int("port", cfg.HTTP.Port),
		slog.Duration("read_timeout", cfg.HTTP.ReadTimeout),
		slog.Duration("write_timeout", cfg.HTTP.WriteTimeout))

	router, err := http_controllers.NewRouter(cfg, log)
	if err != nil {
		log.Error("failed to create router", logger.Err(err))
		return nil, err
	}
	log.Debug("router initialized successfully")

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	log.Info("HTTP server instance created")
	return &HTTPServer{
		server: server,
		log:    log,
		cfg:    cfg,
	}, nil
}

func (s *HTTPServer) Start() error {
	const op = "http_server.HTTPServer.Start"
	log := s.log.With(slog.String("op", op))

	log.Info("starting HTTP server",
		slog.String("addr", s.server.Addr),
		slog.String("network", "tcp"))

	return s.server.ListenAndServe()
}

func (s *HTTPServer) Stop() {
	const op = "http_server.HTTPServer.Stop"
	log := s.log.With(slog.String("op", op))

	log.Info("initiating HTTP server shutdown",
		slog.Duration("timeout", 5*time.Second))
	shutdownStart := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Error("HTTP server shutdown error",
			logger.Err(err),
			slog.Duration("duration", time.Since(shutdownStart)))
		return
	}

	log.Info("HTTP server stopped gracefully",
		slog.Duration("shutdown_duration", time.Since(shutdownStart)))
}
