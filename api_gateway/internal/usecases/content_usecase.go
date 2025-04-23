package usecases

import (
	"api_gateway/internal/domain/models"
	"context"
)

type ContentUsecase interface {
	ProcessContent(userID string, contentType models.ContentType, data string, mimeType string) (*models.Content, error)
	GetContent(ctx context.Context, contentID string) (*models.ContentStatus, error)
	Close() error
}
