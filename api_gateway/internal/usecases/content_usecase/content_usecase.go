package content_usecase

import (
	"api_gateway/internal/config"
	"api_gateway/internal/grpc/storage_client"
	kafka3 "api_gateway/internal/producer/kafka"
	"context"
	"log/slog"
	"time"

	"api_gateway/internal/domain/errors"
	"api_gateway/internal/domain/models"
	"api_gateway/logger"
	"github.com/google/uuid"
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
	return &UseCase{
		producer:    producer,
		log:         log,
		cfg:         cfg,
		storageAddr: cfg.Storage.ServiceAddress,
	}, nil
}

func (uc *UseCase) initStorageClient() error {
	if uc.storage != nil {
		return nil
	}

	client, err := storage_client.NewStorageClient(uc.cfg.Storage.ServiceAddress, 5*time.Second, uc.log)
	if err != nil {
		return err
	}

	uc.storage = client
	return nil
}

func (uc *UseCase) ProcessContent(userID string, contentType models.ContentType, data string) (*models.Content, error) {
	const op = "content_usecase.ProcessContent"
	log := uc.log.With(
		slog.String("op", op),
		slog.String("user_id", userID),
		slog.String("content_type", string(contentType)),
	)

	content := &models.Content{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      contentType,
		Data:      data,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	if err := uc.producer.ProduceContent(content); err != nil {
		log.Error("failed to produce content", logger.Err(err))
		return nil, errors.ErrInternalServer
	}

	log.Info("content processed successfully", slog.String("content_id", content.ID))
	return content, nil
}

func (uc *UseCase) GetContentStatus(ctx context.Context, contentID string) (*models.ContentStatus, error) {
	if err := uc.initStorageClient(); err != nil {
		return nil, err
	}

	return uc.storage.GetContentStatus(ctx, contentID)
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
