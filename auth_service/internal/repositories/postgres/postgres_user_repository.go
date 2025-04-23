package postgres

import (
	"auth_service/internal/config"
	e "auth_service/internal/domain/errors"
	"auth_service/internal/domain/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log/slog"
	"strings"
	"time"
)

const driverName = "postgres"

type UserRepository struct {
	cfg *config.DatabaseConfig
	db  *sql.DB
	log *slog.Logger
}

func NewUserRepository(ctx context.Context, cfg *config.DatabaseConfig, log *slog.Logger) (*UserRepository, error) {
	const op = "postgres.NewUserRepository"
	log = log.With(slog.String("op", op))

	log.Info("connecting to database",
		slog.String("driver", driverName),
		slog.String("dsn_mask", maskDSN(cfg.DSN)),
	)

	startTime := time.Now()
	defer func() {
		log.Info("database connection completed",
			slog.Duration("duration", time.Since(startTime)))
	}()

	db, err := sql.Open(driverName, cfg.DSN)
	if err != nil {
		log.Error("failed to open database connection",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err = db.PingContext(ctx); err != nil {
		log.Error("database ping failed",
			slog.Any("error", err),
			slog.String("driver", driverName),
		)
		_ = db.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	log.Info("database connection established successfully")
	return &UserRepository{
		cfg: cfg,
		db:  db,
		log: log,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) (string, error) {
	const op = "postgres.UserRepository.Create"
	log := r.log.With(
		slog.String("op", op),
	)

	log.Info("creating new user")
	startTime := time.Now()

	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	query, args, err := sq.Insert("users").
		Columns("id", "email", "password", "name", "created_at", "updated_at").
		Values(user.ID, user.Email, user.Password, user.Name, user.CreatedAt, user.UpdatedAt).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Error("failed to build SQL query",
			slog.Any("error", err),
		)
		return "", e.ErrInternalServer
	}

	log.Debug("executing SQL query",
		slog.String("query", query),
		slog.Any("args", []interface{}{
			user.ID,
			user.Email,
			"***",
			user.Name,
			user.CreatedAt,
			user.UpdatedAt,
		}),
		slog.Duration("duration", time.Since(startTime)),
	)

	err = r.db.QueryRowContext(ctx, query, args...).Scan(&user.ID)
	if err != nil {
		if isDuplicateKeyError(err) {
			log.Warn("user already exists",
				slog.Any("error", err),
			)
			return "", e.ErrUserAlreadyExists
		}
		log.Error("failed to create user",
			slog.Any("error", err),
		)
		return "", e.ErrInternalServer
	}

	log.Info("user created successfully",
		slog.String("user_id", user.ID),
		slog.Duration("duration", time.Since(startTime)),
	)
	return user.ID, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "postgres.UserRepository.GetByEmail"
	log := r.log.With(
		slog.String("op", op),
	)

	log.Debug("getting user by email")
	startTime := time.Now()

	query, args, err := sq.
		Select("id", "email", "password", "name", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"email": email}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Error("failed to build SQL query",
			slog.Any("error", err),
		)
		return nil, e.ErrInternalServer
	}

	log.Debug("executing SQL query",
		slog.String("query", query),
		slog.Any("args", args),
		slog.Duration("duration", time.Since(startTime)))

	var user models.User
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Debug("user not found",
				slog.Duration("duration", time.Since(startTime)))
			return nil, e.ErrUserNotFound
		}
		log.Error("failed to get user by email",
			slog.Any("error", err),
			slog.Duration("duration", time.Since(startTime)))
		return nil, e.ErrInternalServer
	}

	log.Debug("user found",
		slog.String("user_id", user.ID),
		slog.Duration("duration", time.Since(startTime)),
	)
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	const op = "postgres.UserRepository.GetByID"
	log := r.log.With(
		slog.String("op", op),
		slog.String("user_id", id),
	)

	log.Debug("getting user by ID")
	startTime := time.Now()

	query, args, err := sq.
		Select("id", "email", "password", "name", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Error("failed to build SQL query",
			slog.Any("error", err),
		)
		return nil, e.ErrInternalServer
	}

	log.Debug("executing SQL query",
		slog.String("query", query),
		slog.Any("args", args),
		slog.Duration("duration", time.Since(startTime)),
	)

	var user models.User
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Debug("user not found",
				slog.Duration("duration", time.Since(startTime)))
			return nil, e.ErrUserNotFound
		}
		log.Error("failed to get user by ID",
			slog.Any("error", err),
			slog.Duration("duration", time.Since(startTime)))
		return nil, e.ErrInternalServer
	}

	log.Debug("user found",
		slog.String("email", user.Email),
		slog.Duration("duration", time.Since(startTime)),
	)
	return &user, nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "23505")
}

func maskDSN(dsn string) string {
	if strings.Contains(dsn, "@") {
		parts := strings.Split(dsn, "@")
		if len(parts) > 1 {
			return "***@" + parts[1]
		}
	}
	return "***"
}
