package grpc

import (
	"api_gateway/internal/domain/models"
	"context"
)

type StorageClient interface {
	GetContentStatus(ctx context.Context, contentID string) (*models.ContentStatus, error)
	Close() error
}
