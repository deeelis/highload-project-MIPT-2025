package repos

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"storage_service/internal/config"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"storage_service/internal/domain/models"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type PostgresContentRepository struct {
	cfg *config.RepoConfig
	db  *sql.DB
	log *slog.Logger
}

const driverName = "postgres"

func NewPostgresContentRepository(ctx context.Context, cfg *config.RepoConfig, log *slog.Logger) (*PostgresContentRepository, error) {
	log.Info("connecting to database",
		slog.String("driver", driverName),
		slog.String("dsn_mask", maskDSN(cfg.DSN)), // Функция для маскирования чувствительных данных
	)

	db, err := sql.Open(driverName, cfg.DSN)
	if err != nil {
		log.Error("failed to open database connection",
			slog.Any("error", err),
			slog.String("driver", driverName),
		)
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
	return &PostgresContentRepository{cfg: cfg,
		db:  db,
		log: log}, nil
}

func (r *PostgresContentRepository) CreateContent(ctx context.Context, content *models.Content) error {
	metadataJSON, err := json.Marshal(content.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query, args, err := psql.Insert("content").
		Columns("id", "type", "status", "metadata", "created_at", "updated_at").
		Values(content.ID, content.Type, content.Status, metadataJSON, content.CreatedAt, content.UpdatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PostgresContentRepository) GetContent(ctx context.Context, id string) (*models.Content, error) {
	query, args, err := psql.Select("id", "type", "status", "metadata", "created_at", "updated_at").
		From("content").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var content models.Content
	var metadataJSON []byte

	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&content.ID,
		&content.Type,
		&content.Status,
		&metadataJSON,
		&content.CreatedAt,
		&content.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("content not found")
		}
		return nil, err
	}

	if err := json.Unmarshal(metadataJSON, &content.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &content, nil
}

func (r *PostgresContentRepository) UpdateContentStatus(ctx context.Context, id string, status models.ProcessingStatus) error {
	query, args, err := psql.Update("content").
		Set("status", status).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PostgresContentRepository) UpdateTextContent(ctx context.Context, content *models.TextContent) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	r.log.Debug(content.ID)
	metadataJSON, err := json.Marshal(content.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	updateContentQuery, updateArgs, err := psql.Update("content").
		Set("status", content.Status).
		Set("metadata", metadataJSON).
		Set("updated_at", content.UpdatedAt).
		Where(sq.Eq{"id": content.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update content query: %w", err)
	}

	if _, err = tx.ExecContext(ctx, updateContentQuery, updateArgs...); err != nil {
		r.log.Error("first exec " + err.Error())
		return err
	}

	r.log.Debug(content.ID)
	upsertTextQuery, textArgs, err := psql.Insert("text_content").
		Columns("content_id", "original_text").
		Values(content.ID, content.OriginalText).
		Suffix("ON CONFLICT (content_id) DO UPDATE SET original_text = ?", content.OriginalText).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build upsert text query: %w", err)
	}

	if _, err = tx.ExecContext(ctx, upsertTextQuery, textArgs...); err != nil {
		r.log.Error("second exec " + err.Error())
		return err
	}

	return tx.Commit()
}

func (r *PostgresContentRepository) UpdateImageContent(ctx context.Context, content *models.ImageContent) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	metadataJSON, err := json.Marshal(content.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	updateContentQuery, updateArgs, err := psql.Update("content").
		Set("status", content.Status).
		Set("metadata", metadataJSON).
		Set("updated_at", content.UpdatedAt).
		Where(sq.Eq{"id": content.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update content query: %w", err)
	}

	if _, err = tx.ExecContext(ctx, updateContentQuery, updateArgs...); err != nil {
		return err
	}

	upsertImageQuery, imageArgs, err := psql.Insert("image_content").
		Columns("content_id", "s3_key").
		Values(content.ID, content.S3Key).
		Suffix("ON CONFLICT (content_id) DO UPDATE SET s3_key = ?", content.S3Key).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build upsert image query: %w", err)
	}

	if _, err = tx.ExecContext(ctx, upsertImageQuery, imageArgs...); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresContentRepository) GetImageContent(ctx context.Context, id string) (*models.ImageContent, error) {
	content, err := r.GetContent(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	query, args, err := psql.Select("content_id", "s3_key").
		From("image_content").
		Where(sq.Eq{"content_id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var imageContent models.ImageContent
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&imageContent.ID,
		&imageContent.S3Key,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("image content not found")
		}
		return nil, err
	}

	imageContent.Type = content.Type
	imageContent.Status = content.Status
	imageContent.Metadata = content.Metadata
	imageContent.CreatedAt = content.CreatedAt
	imageContent.UpdatedAt = content.UpdatedAt

	return &imageContent, nil
}

func (r *PostgresContentRepository) GetTextContent(ctx context.Context, id string) (*models.TextContent, error) {
	r.log.Debug(id)
	content, err := r.GetContent(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	r.log.Debug(content.ID)
	r.log.Debug(string(content.Status))
	query, args, err := psql.Select("content_id", "original_text").
		From("text_content").
		Where(sq.Eq{"content_id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var textContent models.TextContent
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&textContent.ID,
		&textContent.OriginalText,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("text content not found")
		}
		return nil, err
	}

	textContent.Type = content.Type
	textContent.Status = content.Status
	textContent.Metadata = content.Metadata
	textContent.CreatedAt = content.CreatedAt
	textContent.UpdatedAt = content.UpdatedAt

	return &textContent, nil
}

func (r *PostgresContentRepository) CreateContentRecord(ctx context.Context, content *models.Content) error {
	metadata, _ := json.Marshal(make(map[string]string))

	query, args, err := psql.Insert("content").
		Columns("id", "type", "status", "created_at", "updated_at", "metadata").
		Values(content.ID, content.Type, content.Status, content.CreatedAt, content.UpdatedAt, metadata).
		Suffix("ON CONFLICT (id) DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
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
