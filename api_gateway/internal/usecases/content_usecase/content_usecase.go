package content_usecase

import (
	"api_gateway/internal/config"
	"api_gateway/internal/domain/errors"
	"api_gateway/internal/domain/models"
	"api_gateway/internal/grpc/storage_client"
	kafka3 "api_gateway/internal/producer/kafka"
	"api_gateway/logger"
	"context"
	"github.com/google/uuid"
	"log/slog"
)

type UseCase struct {
	cfg         *config.Config
	producer    *kafka3.Producer
	storage     *storage_client.StorageClient
	log         *slog.Logger
	storageAddr string
}

func NewContentUseCase(ctx context.Context, cfg *config.Config, log *slog.Logger) (*UseCase, error) {
	producer, err := kafka3.NewProducer(ctx, cfg.Kafka, log)
	if err != nil {
		return nil, err
	}
	client, err := storage_client.NewStorageClient(cfg.Storage, log)
	if err != nil {
		return nil, err
	}
	return &UseCase{
		producer:    producer,
		log:         log,
		storage:     client,
		cfg:         cfg,
		storageAddr: cfg.Storage.ServiceAddress,
	}, nil
}

func (uc *UseCase) ProcessContent(userID string, contentType models.ContentType, data string, mimeType string) (*models.Content, error) {
	ctx := context.Background()
	const op = "content_usecase.ProcessContent"
	log := uc.log.With(
		slog.String("op", op),
		slog.String("user_id", userID),
		slog.String("content_type", string(contentType)),
	)

	content := &models.Content{
		ID:       uuid.New().String(),
		UserID:   userID,
		Type:     contentType,
		Data:     data,
		DataType: mimeType,
	}

	if err := uc.storage.RegisterContent(ctx, content.ID, string(content.Type)); err != nil {
		return nil, errors.ErrInternalServer
	}

	if err := uc.producer.ProduceContent(content); err != nil {
		log.Error("failed to produce content", logger.Err(err))
		return nil, errors.ErrInternalServer
	}

	log.Info("content processed successfully", slog.String("content_id", content.ID))
	return content, nil
}

func (uc *UseCase) GetContent(ctx context.Context, contentID string) (*models.ContentStatus, error) {
	if contentID == "" {
		return nil, errors.ErrInvalidCredentials
	}
	return uc.storage.GetContent(ctx, contentID)
}

func (uc *UseCase) Close() error {
	var err error
	if closeErr := uc.producer.Close(); closeErr != nil {
		uc.log.Error("failed to close producer", logger.Err(closeErr))
		err = closeErr
	}

	if uc.storage != nil {
		if closeErr := uc.storage.Close(); closeErr != nil {
			uc.log.Error("failed to close storage client", logger.Err(closeErr))
			err = closeErr
		}
	}

	return err
}
