package content_usecase

import (
	"api_gateway/internal/config"
	"api_gateway/internal/domain/errors"
	"api_gateway/internal/domain/models"
	"api_gateway/internal/grpc"
	"api_gateway/internal/grpc/storage_client"
	kafka3 "api_gateway/internal/producer/kafka"
	"api_gateway/logger"
	"context"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type ContentUseCase struct {
	cfg      *config.Config
	producer *kafka3.Producer
	storage  grpc.StorageClient
	log      *slog.Logger
}

func NewContentUseCase(ctx context.Context, cfg *config.Config, log *slog.Logger) (*ContentUseCase, error) {
	const op = "content_usecase.New"
	log = log.With(slog.String("op", op))
	log.Info("initializing", slog.String("storage_addr", cfg.Storage.ServiceAddress))

	start := time.Now()
	defer func() {
		log.Info("initialization completed", slog.Duration("duration", time.Since(start)))
	}()
	producer, err := kafka3.NewProducer(ctx, cfg.Kafka, log)
	if err != nil {
		log.Error("kafka producer init failed", logger.Err(err))
		return nil, err
	}

	client, err := storage_client.NewStorageClient(cfg.Storage, log)

	if err != nil {
		log.Error("storage client init failed", logger.Err(err))
		return nil, err
	}
	return &ContentUseCase{
		producer: producer,
		log:      log,
		storage:  client,
		cfg:      cfg,
	}, nil
}

func (uc *ContentUseCase) ProcessContent(userID string, contentType models.ContentType, data string, mimeType string) (*models.Content, error) {
	const op = "content_usecase.ProcessContent"
	log := uc.log.With(
		slog.String("op", op),
		slog.String("user_id", userID),
		slog.String("content_type", string(contentType)),
		slog.String("mime_type", mimeType),
	)

	start := time.Now()

	content := &models.Content{
		ID:       uuid.New().String(),
		UserID:   userID,
		Type:     contentType,
		Data:     data,
		DataType: mimeType,
	}

	log.Debug("registering content", slog.String("content_id", content.ID))
	if err := uc.storage.RegisterContent(context.Background(), content.ID, string(content.Type)); err != nil {
		log.Error("content registration failed",
			logger.Err(err),
			slog.Duration("duration", time.Since(start)))
		return nil, errors.ErrInternalServer
	}

	log.Debug("producing content",
		slog.String("content_id", content.ID),
		slog.Int("data_size", len(data)))
	if err := uc.producer.ProduceContent(content); err != nil {
		log.Error("content production failed",
			logger.Err(err),
			slog.Duration("duration", time.Since(start)))
		return nil, errors.ErrInternalServer
	}

	log.Info("content processed",
		slog.String("content_id", content.ID),
		slog.Duration("duration", time.Since(start)))
	return content, nil
}

func (uc *ContentUseCase) GetContent(ctx context.Context, contentID string) (*models.ContentStatus, error) {
	const op = "content_usecase.GetContent"
	log := uc.log.With(
		slog.String("op", op),
		slog.String("content_id", contentID),
	)

	if contentID == "" {
		log.Warn("empty content id")
		return nil, errors.ErrInvalidCredentials
	}
	start := time.Now()
	status, err := uc.storage.GetContent(ctx, contentID)
	if err != nil {
		log.Error("content fetch failed",
			logger.Err(err),
			slog.Duration("duration", time.Since(start)))
		return nil, err
	}

	log.Info("content fetched",
		slog.String("status", status.Status),
		slog.Duration("duration", time.Since(start)))
	return status, nil
}

func (uc *ContentUseCase) Close() error {
	const op = "content_usecase.Close"
	log := uc.log.With(slog.String("op", op))
	log.Info("closing resources")

	var err error
	start := time.Now()

	if closeErr := uc.producer.Close(); closeErr != nil {
		log.Error("producer close failed", logger.Err(closeErr))
		err = closeErr
	} else {
		log.Debug("producer closed")
	}

	if uc.storage != nil {
		if closeErr := uc.storage.Close(); closeErr != nil {
			log.Error("storage client close failed", logger.Err(closeErr))
			err = closeErr
		} else {
			log.Debug("storage client closed")
		}
	}

	log.Info("resources closed",
		slog.Duration("duration", time.Since(start)),
		slog.Bool("success", err == nil))

	return err
}
