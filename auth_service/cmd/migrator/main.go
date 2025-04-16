package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
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
			log.Fatal("database connection URL is required (use -dsn flag or DB_URL env var)")
		}
	}

	if !strings.HasPrefix(dbURL, "postgres://") && !strings.HasPrefix(dbURL, "postgresql://") {
		dbURL = "postgres://" + dbURL
	}

	if migrationsPath == "" {
		log.Fatal("migrations path is required (use -migrations-path flag)")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("%s", dbURL),
	)
	if err != nil {
		log.Fatalf("failed to initialize migrator: %v", err)
	}
	defer m.Close()

	switch action {
	case "up":
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("no migrations to apply")
				return
			}
			log.Fatalf("failed to apply migrations: %v", err)
		}
		log.Println("migrations applied successfully")

	case "down":
		if err := m.Down(); err != nil {
			log.Fatalf("failed to rollback migrations: %v", err)
		}
		log.Println("migrations rolled back successfully")

	case "force":
		version := flag.Arg(0)
		if version == "" {
			log.Fatal("version is required for force action")
		}
		v, err := strconv.Atoi(version)
		if err != nil {
			log.Fatalf("invalid version number: %v", err)
		}
		if err := m.Force(v); err != nil {
			log.Fatalf("failed to force migration: %v", err)
		}
		log.Printf("forced migration to version %d\n", v)

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("failed to get migration version: %v", err)
		}
		log.Printf("current migration version: %d (dirty: %v)\n", version, dirty)

	default:
		log.Fatalf("unknown action: %s (available: up, down, force, version)", action)
	}
}
