package repositories

import (
	"context"
	"storage_service/internal/domain/models"
)

type ContentRepository interface {
	CreateContentRecord(ctx context.Context, content *models.Content) error
	CreateContent(ctx context.Context, content *models.Content) error
	GetContent(ctx context.Context, id string) (*models.Content, error)
	UpdateContentStatus(ctx context.Context, id string, status models.ProcessingStatus) error
	UpdateTextContent(ctx context.Context, content *models.TextContent) error
	UpdateImageContent(ctx context.Context, content *models.ImageContent) error
	GetTextContent(ctx context.Context, id string) (*models.TextContent, error)
	GetImageContent(ctx context.Context, id string) (*models.ImageContent, error)
}

type CacheRepository interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl int) error
	Delete(ctx context.Context, key string) error
}

type StorageUsecase interface {
	GetContent(ctx context.Context, id string) (*models.Content, error)
	GetTextContent(ctx context.Context, id string) (*models.TextContent, error)
	GetImageContent(ctx context.Context, id string) (*models.ImageContent, error)
	StoreTextAnalysis(ctx context.Context, content *models.TextContent) error
	StoreImageAnalysis(ctx context.Context, content *models.ImageContent) error
	ProcessImageMessage(ctx context.Context, m *models.ImageMessage) error
	ProcessTextMessage(ctx context.Context, m *models.TextMessage) error
	CreateContentRecord(ctx context.Context, contentType models.ContentType, contentID string) error
}

type ImageStorage interface {
	StoreImage(ctx context.Context, data []byte) (string, error)
	GetImageURL(ctx context.Context, key string) (string, error)
}
