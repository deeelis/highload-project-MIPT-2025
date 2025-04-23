package storage_client

import (
	"api_gateway/internal/config"
	"api_gateway/internal/domain/models"
	"api_gateway/logger"
	"context"
	"fmt"
	storagepb "github.com/deeelis/storage-protos/gen/go/storage"
	"google.golang.org/grpc/status"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type StorageClient struct {
	client storagepb.StorageServiceClient
	conn   *grpc.ClientConn
	cfg    *config.StorageConfig
	log    *slog.Logger
}

func NewStorageClient(cfg *config.StorageConfig, log *slog.Logger) (*StorageClient, error) {
	const op = "storage_client.NewStorageClient"
	log = log.With(slog.String("op", op))
	log.Info("initializing storage client",
		slog.String("address", cfg.ServiceAddress),
		slog.Duration("timeout", cfg.Timeout))

	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		cfg.ServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Error("failed to connect to storage service",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return nil, fmt.Errorf("failed to connect to storage service: %w", err)
	}

	log.Info("storage client initialized successfully",
		slog.Duration("duration", time.Since(startTime)))

	return &StorageClient{
		client: storagepb.NewStorageServiceClient(conn),
		conn:   conn,
		cfg:    cfg,
		log:    log,
	}, nil
}

func (c *StorageClient) GetContent(ctx context.Context, contentID string) (*models.ContentStatus, error) {
	const op = "storage_client.GetContent"
	log := c.log.With(
		slog.String("op", op),
		slog.String("content_id", contentID),
	)

	log.Info("getting content from storage")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	resp, err := c.client.GetContent(ctx, &storagepb.ContentRequest{
		ContentId: contentID,
	})
	if err != nil {
		grpcStatus, _ := status.FromError(err)
		log.Error("failed to get content",
			logger.Err(err),
			slog.String("grpc_code", grpcStatus.Code().String()),
			slog.Duration("duration", time.Since(startTime)))
		return nil, fmt.Errorf("failed to get content status: %w", err)
	}

	contentStatus := &models.ContentStatus{
		ID:     resp.ContentId,
		Type:   resp.Type.String(),
		Status: resp.Status.String(),
	}

	switch resp.Type {
	case storagepb.ContentType_TEXT:
		if text := resp.GetText(); text != nil {
			contentStatus.OriginalContent = text.OriginalText
			contentStatus.Analysis = convertMetadata(text.AnalysisMetadata)
			log.Debug("retrieved text content",
				slog.Int("text_length", len(text.OriginalText)),
				slog.Int("metadata_items", len(text.AnalysisMetadata)))
		}
	case storagepb.ContentType_IMAGE:
		if image := resp.GetImage(); image != nil {
			contentStatus.OriginalContent = image.ImageUrl
			contentStatus.Analysis = convertMetadata(image.AnalysisMetadata)
			log.Debug("retrieved image content",
				slog.String("image_url", image.ImageUrl),
				slog.Int("metadata_items", len(image.AnalysisMetadata)))
		}
	}

	log.Info("content retrieved successfully",
		slog.String("content_type", resp.Type.String()),
		slog.String("status", resp.Status.String()),
		slog.Duration("duration", time.Since(startTime)))
	return contentStatus, nil
}

func (c *StorageClient) RegisterContent(ctx context.Context, contentID string, contentType string) error {
	const op = "storage_client.RegisterContent"
	log := c.log.With(
		slog.String("op", op),
		slog.String("content_id", contentID),
		slog.String("content_type", contentType),
	)

	log.Info("registering content in storage")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	var ct storagepb.ContentType
	switch contentType {
	case "text":
		ct = storagepb.ContentType_TEXT
	case "image":
		ct = storagepb.ContentType_IMAGE
	default:
		log.Error("invalid content type provided",
			slog.Duration("duration", time.Since(startTime)))
		return fmt.Errorf("invalid content type: %s", contentType)
	}

	_, err := c.client.RegisterContent(ctx, &storagepb.RegisterContentRequest{
		ContentId: contentID,
		Type:      ct,
	})
	if err != nil {
		grpcStatus, _ := status.FromError(err)
		log.Error("failed to register content",
			logger.Err(err),
			slog.String("grpc_code", grpcStatus.Code().String()),
			slog.Duration("duration", time.Since(startTime)))
		return fmt.Errorf("failed to register content: %w", err)
	}

	log.Info("content registered successfully",
		slog.Duration("duration", time.Since(startTime)))

	return nil
}

func convertMetadata(metadata map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range metadata {
		result[k] = v
	}
	return result
}

func (c *StorageClient) Close() error {
	const op = "storage_client.Close"
	log := c.log.With(slog.String("op", op))

	log.Info("closing storage client connection")
	startTime := time.Now()
	err := c.conn.Close()
	if err != nil {
		log.Error("failed to close connection",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return err
	}

	log.Info("connection closed successfully",
		slog.Duration("duration", time.Since(startTime)))

	return nil
}
