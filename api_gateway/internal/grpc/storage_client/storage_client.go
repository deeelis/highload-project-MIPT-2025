package storage_client

import (
	"api_gateway/internal/config"
	"api_gateway/internal/domain/models"
	"context"
	"fmt"
	storagepb "github.com/deeelis/storage-protos/gen/go/storage"
	"log/slog"

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
	conn, err := grpc.Dial(
		cfg.ServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to storage service: %w", err)
	}

	return &StorageClient{
		client: storagepb.NewStorageServiceClient(conn),
		conn:   conn,
		cfg:    cfg,
		log:    log,
	}, nil
}

func (c *StorageClient) GetContent(ctx context.Context, contentID string) (*models.ContentStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	resp, err := c.client.GetContent(ctx, &storagepb.ContentRequest{
		ContentId: contentID,
	})
	if err != nil {
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
		}
	case storagepb.ContentType_IMAGE:
		if image := resp.GetImage(); image != nil {
			contentStatus.OriginalContent = image.ImageUrl
			contentStatus.Analysis = convertMetadata(image.AnalysisMetadata)
		}
	}

	return contentStatus, nil
}

func (c *StorageClient) RegisterContent(ctx context.Context, contentID string, contentType string) error {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	var ct storagepb.ContentType
	switch contentType {
	case "TEXT":
		ct = storagepb.ContentType_TEXT
	case "IMAGE":
		ct = storagepb.ContentType_IMAGE
	default:
		return fmt.Errorf("invalid content type: %s", contentType)
	}

	_, err := c.client.RegisterContent(ctx, &storagepb.RegisterContentRequest{
		ContentId: contentID,
		Type:      ct,
	})
	if err != nil {
		return fmt.Errorf("failed to register content: %w", err)
	}

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
	return c.conn.Close()
}
