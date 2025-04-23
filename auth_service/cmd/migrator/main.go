package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	startTime := time.Now()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	var (
		dbURL           string
		migrationsPath  string
		migrationsTable string
		action          string
	)

	flag.StringVar(&dbURL, "dsn", "", "PostgreSQL connection URL (format: postgres://user:pass@host:port/db?sslmode=disable)")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations directory")
	flag.StringVar(&migrationsTable, "migrations-table", "schema_migrations", "name of migrations table")
	flag.StringVar(&action, "action", "up", "migration action (up, down, force, version)")
	flag.Parse()

	if dbURL == "" {
		dbURL = os.Getenv("DB_URL")
		if dbURL == "" {
			logger.Error("database connection URL is required")
			os.Exit(1)
		}
	}

	if !strings.HasPrefix(dbURL, "postgres://") && !strings.HasPrefix(dbURL, "postgresql://") {
		dbURL = "postgres://" + dbURL
	}

	loggableDBURL := dbURL
	if atIndex := strings.Index(dbURL, "@"); atIndex > 0 {
		loggableDBURL = "postgres://*****" + dbURL[atIndex:]
	}

	if migrationsPath == "" {
		logger.Error("migrations path is required")
		os.Exit(1)
	}

	logger.Info("initializing migrator",
		slog.String("migrations_path", migrationsPath),
		slog.String("db_url", loggableDBURL))

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("%s", dbURL),
	)
	if err != nil {
		logger.Error("failed to initialize migrator",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if _, err := m.Close(); err != nil {
			logger.Error("failed to close migrator",
				slog.String("error", err.Error()))
		}
	}()

	switch action {
	case "up":
		logger.Info("applying migrations")
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				logger.Info("no migrations to apply")
				return
			}
			logger.Error("failed to apply migrations",
				slog.String("error", err.Error()))
			os.Exit(1)
		}
		logger.Info("migrations applied successfully",
			slog.Duration("duration", time.Since(startTime)))
	case "down":
		logger.Info("rolling back migrations")
		if err := m.Down(); err != nil {
			logger.Error("failed to rollback migrations",
				slog.String("error", err.Error()))
			os.Exit(1)
		}
		logger.Info("migrations rolled back successfully",
			slog.Duration("duration", time.Since(startTime)))
	case "force":
		version := flag.Arg(0)
		if version == "" {
			logger.Error("version is required for force action")
			os.Exit(1)
		}
		v, err := strconv.Atoi(version)
		if err != nil {
			logger.Error("invalid version number",
				slog.String("version", version),
				slog.String("error", err.Error()))
			os.Exit(1)
		}
		logger.Info("forcing migration version",
			slog.Int("version", v))
		if err := m.Force(v); err != nil {
			logger.Error("failed to force migration",
				slog.Int("version", v),
				slog.String("error", err.Error()))
			os.Exit(1)
		}
		logger.Info("migration forced successfully",
			slog.Int("version", v),
			slog.Duration("duration", time.Since(startTime)))
	case "version":
		logger.Info("checking migration version")
		version, dirty, err := m.Version()
		if err != nil {
			logger.Error("failed to get migration version",
				slog.String("error", err.Error()))
			os.Exit(1)
		}
		logger.Info("current migration version",
			slog.Int("version", int(version)),
			slog.Bool("dirty", dirty),
			slog.Duration("duration", time.Since(startTime)))
	default:
		logger.Error("unknown action",
			slog.String("action", action))
		os.Exit(1)
	}
}
