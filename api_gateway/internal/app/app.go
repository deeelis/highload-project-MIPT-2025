package app

import (
	"api_gateway/internal/app/http_server"
	"api_gateway/internal/config"
	"api_gateway/logger"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type App struct {
	cfg    *config.Config
	log    *slog.Logger
	server *http_server.HTTPServer
}

func NewApp(log *slog.Logger, cfg *config.Config) (*App, error) {
	const op = "app.NewApp"
	log = log.With(slog.String("op", op))

	log.Info("initializing application",
		slog.String("env", cfg.Env),
		slog.Int("http_port", cfg.HTTP.Port),
		slog.Duration("http_read_timeout", cfg.HTTP.ReadTimeout),
		slog.Duration("http_write_timeout", cfg.HTTP.WriteTimeout))

	server, err := http_server.NewHTTPServer(log, cfg)
	if err != nil {
		log.Error("failed to create HTTP server", logger.Err(err))
		return nil, errors.Join(err, errors.New("http server creation failed"))
	}

	log.Info("application initialized successfully")
	return &App{
		cfg:    cfg,
		log:    log,
		server: server,
	}, nil
}

func (a *App) Run() error {
	const op = "app.Run"
	log := a.log.With(slog.String("op", op))
	startTime := time.Now()

	log.Info("starting application services")
	defer log.Info("application services stopped")

	if err := a.server.Start(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed unexpectedly",
				logger.Err(err),
				slog.Duration("uptime", time.Since(startTime)))
			return errors.Join(err, errors.New("server run failed"))
		}
		log.Info("server closed normally")
	}
	return a.server.Start()
}

func (a *App) Stop() {
	const op = "app.Stop"
	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("initiating application shutdown")
	shutdownStart := time.Now()

	a.server.Stop()

	log.Info("application shutdown completed",
		slog.Duration("shutdown_duration", time.Since(shutdownStart)))
}
