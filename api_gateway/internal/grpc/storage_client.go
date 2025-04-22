package grpc

import (
	"api_gateway/internal/domain/models"
	"context"
)

type StorageClient interface {
	GetContentStatus(ctx context.Context, contentID string) (*models.ContentStatus, error)
	RegisterContent(ctx context.Context, contentID string, contentType string) error
	Close() error
}
