package s3

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	config2 "storage_service/internal/config"
	"time"
)

type S3ImageStorage struct {
	client *S3Client
}

func NewS3ImageStorage(cfg *config2.S3Config, log *slog.Logger) (*S3ImageStorage, error) {
	client, err := NewS3Client(cfg)
	if err != nil {
		log.Info(cfg.Bucket, cfg.Endpoint, cfg.SecretKey, cfg.AccessKey, cfg.Region)
		return nil, err
	}
	return &S3ImageStorage{
		client: client,
	}, nil
}

func (s *S3ImageStorage) StoreImage(ctx context.Context, data []byte) (string, error) {
	key := fmt.Sprintf("images/%d/%s", time.Now().Unix(), uuid.New().String())

	err := s.client.UploadImage(ctx, s.client.bucket, key)
	if err != nil {
		return "", fmt.Errorf("failed to upload image to S3: %w", err)
	}

	return key, nil
}

func (s *S3ImageStorage) GetImageURL(ctx context.Context, key string) (string, error) {
	imageURL := s.client.GetImageURL(key)
	return imageURL, nil
}
