package app

import (
	"api_gateway/internal/app/http_server"
	"api_gateway/internal/config"
	"log/slog"
)

type App struct {
	cfg    *config.Config
	log    *slog.Logger
	server *http_server.HTTPServer
}

func NewApp(log *slog.Logger, cfg *config.Config) (*App, error) {
	server, err := http_server.NewHTTPServer(log, cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		cfg:    cfg,
		log:    log,
		server: server,
	}, nil
}

func (a *App) Run() error {
	a.log.Info("starting HTTP server", slog.Int("port", a.cfg.HTTP.Port))
	return a.server.Start()
}

func (a *App) Stop() {
	a.log.Info("stopping HTTP server")
	a.server.Stop()
}
