package services

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"storage_service/internal/config"
	"storage_service/internal/domain/models"
	"storage_service/internal/domain/repositories"
	"storage_service/internal/repositories/cache"
	repos "storage_service/internal/repositories/repos/postgres"
	"storage_service/internal/repositories/s3"
	"time"
)

type storageUsecase struct {
	cfg         *config.Config
	contentRepo repositories.ContentRepository
	cacheRepo   repositories.CacheRepository
	imageStore  repositories.ImageStorage
	log         *slog.Logger
}

func NewStorageUsecase(ctx context.Context, cfg *config.Config, log *slog.Logger) (repositories.StorageUsecase, error) {
	contentRepo, err := repos.NewPostgresContentRepository(ctx, &cfg.Repo, log)
	if err != nil {
		return nil, err
	}

	cacheRepo, err := cache.NewRedisCache(&cfg.Cache, log)
	if err != nil {
		return nil, err
	}

	imageStore, err := s3.NewS3ImageStorage(&cfg.S3, log)
	if err != nil {
		return nil, err
	}

	return &storageUsecase{
		contentRepo: contentRepo,
		cacheRepo:   cacheRepo,
		imageStore:  imageStore,
		log:         log,
		cfg:         cfg,
	}, nil
}

func (s *storageUsecase) GetTextContent(ctx context.Context, id string) (*models.TextContent, error) {
	var content *models.Content
	if cached, err := s.cacheRepo.Get(ctx, id); err == nil {
		c, ok := cached.(*models.Content)
		if !ok {
			content, err = s.contentRepo.GetContent(ctx, id)
			if err != nil {
				return nil, err
			}
		} else {
			content = c
		}
	} else {
		content, err = s.contentRepo.GetContent(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	//content, err := s.contentRepo.GetContent(ctx, id)
	//if err != nil {
	//	return nil, err
	//}
	s.log.Debug(content.ID)
	s.log.Debug(string(content.Status))

	var err error
	var textContent *models.TextContent
	if content.Status == models.StatusCompleted {
		textContent, err = s.contentRepo.GetTextContent(ctx, content.ID)
		if err != nil {
			return nil, err
		}
		s.log.Debug("get text")
		s.log.Debug(string(textContent.Status))
		return textContent, nil
	} else {
		textContent = &models.TextContent{
			Content: models.Content{
				ID:        content.ID,
				UserID:    content.UserID,
				Type:      content.Type,
				Status:    content.Status,
				CreatedAt: content.CreatedAt,
				UpdatedAt: content.UpdatedAt,
				Metadata:  make(map[string]string),
			},
			OriginalText: "",
		}
	}

	if err := s.cacheRepo.Set(ctx, id, content, 300); err != nil {
		log.Println(err.Error())
	}

	return textContent, nil
}

func (s *storageUsecase) GetImageContent(ctx context.Context, id string) (*models.ImageContent, error) {
	var content *models.Content
	if cached, err := s.cacheRepo.Get(ctx, id); err == nil {
		c, ok := cached.(*models.Content)
		if !ok {
			content, err = s.contentRepo.GetContent(ctx, id)
			if err != nil {
				return nil, err
			}
		} else {
			content = c
		}
	} else {
		content, err = s.contentRepo.GetContent(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	var err error
	var imageContent *models.ImageContent
	if content.Status == models.StatusCompleted {
		imageContent, err = s.contentRepo.GetImageContent(ctx, content.ID)
		if err != nil {
			return nil, err
		}
		return imageContent, nil
	} else {
		imageContent = &models.ImageContent{
			Content: models.Content{
				ID:        content.ID,
				UserID:    content.UserID,
				Type:      content.Type,
				Status:    content.Status,
				CreatedAt: content.CreatedAt,
				UpdatedAt: content.UpdatedAt,
				Metadata:  make(map[string]string),
			},
			S3Key: "",
		}
	}

	if err := s.cacheRepo.Set(ctx, id, content, 300); err != nil {
		log.Println(err.Error())
	}

	return imageContent, nil
}

func (s *storageUsecase) GetContent(ctx context.Context, id string) (*models.Content, error) {
	if cached, err := s.cacheRepo.Get(ctx, id); err == nil {
		if content, ok := cached.(*models.Content); ok {
			return content, nil
		}
	}

	content, err := s.contentRepo.GetContent(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.cacheRepo.Set(ctx, id, content, 300); err != nil {
		log.Println(err.Error())
	}

	return content, nil
}

func (s *storageUsecase) StoreTextAnalysis(ctx context.Context, content *models.TextContent) error {
	if content.ID == "" {
		return errors.New("content ID cannot be empty")
	}
	if content.OriginalText == "" {
		return errors.New("original text cannot be empty")
	}

	content.Status = models.StatusCompleted
	content.UpdatedAt = time.Now()

	if err := s.contentRepo.UpdateTextContent(ctx, content); err != nil {
		return err
	}

	if err := s.cacheRepo.Set(ctx, content.ID, content, 300); err != nil {
		log.Println(err.Error())
	}

	return nil
}

func (s *storageUsecase) StoreImageAnalysis(ctx context.Context, content *models.ImageContent) error {
	if content.ID == "" {
		return errors.New("content ID cannot be empty")
	}
	if content.S3Key == "" {
		return errors.New("S3 key cannot be empty")
	}

	content.Status = models.StatusCompleted
	content.UpdatedAt = time.Now()

	if err := s.contentRepo.UpdateImageContent(ctx, content); err != nil {
		return err
	}

	if err := s.cacheRepo.Set(ctx, content.ID, content, 300); err != nil {
	}

	return nil
}

func (s *storageUsecase) ProcessTextMessage(ctx context.Context, msg *models.TextMessage) error {
	content := models.Content{
		ID:        msg.ID,
		Type:      models.ContentTypeText,
		Status:    models.StatusCompleted,
		Metadata:  msg.Analysis,
		UpdatedAt: time.Now(),
	}

	textContent := &models.TextContent{
		Content:      content,
		OriginalText: msg.Content,
	}

	if err := s.contentRepo.UpdateTextContent(ctx, textContent); err != nil {
		return fmt.Errorf("failed to update text content: %w", err)
	}

	if err := s.cacheRepo.Set(ctx, textContent.ID, textContent, 300); err != nil {
		log.Printf("cache set error: %v", err)
	}
	return nil
}

func (s *storageUsecase) ProcessImageMessage(ctx context.Context, msg *models.ImageMessage) error {
	data, err := base64.StdEncoding.DecodeString(msg.Data)
	if err != nil {
		return fmt.Errorf("can't decode image")
	}

	key, err := s.imageStore.StoreImage(ctx, data)
	if err != nil {
		return err
	}

	url, err := s.imageStore.GetImageURL(ctx, key)
	if err != nil {
		return err
	}

	metadata := make(map[string]string)
	metadata["drawings"] = fmt.Sprintf("%f", msg.NsfwScores.Drawings)
	metadata["sexy"] = fmt.Sprintf("%f", msg.NsfwScores.Sexy)
	metadata["porn"] = fmt.Sprintf("%f", msg.NsfwScores.Porn)
	metadata["neutral"] = fmt.Sprintf("%f", msg.NsfwScores.Neutral)
	metadata["hentai"] = fmt.Sprintf("%f", msg.NsfwScores.Hentai)

	content := &models.ImageContent{
		Content: models.Content{
			ID:        msg.ID,
			UserID:    msg.UserID,
			Type:      models.ContentTypeImage,
			Status:    models.StatusCompleted,
			Metadata:  metadata,
			UpdatedAt: time.Now(),
		},
		S3Key: url,
	}

	if err := s.contentRepo.UpdateImageContent(ctx, content); err != nil {
		return fmt.Errorf("failed to update image content: %w", err)
	}

	if err := s.cacheRepo.Set(ctx, content.ID, content, 300); err != nil {
		log.Printf("cache set error: %v", err)
	}
	return nil
}

func (s *storageUsecase) CreateContentRecord(ctx context.Context, contentType models.ContentType, contentID string) error {
	content := &models.Content{
		ID:        contentID,
		Type:      contentType,
		Status:    models.StatusProcessing,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.contentRepo.CreateContentRecord(ctx, content); err != nil {
		return fmt.Errorf("failed to create content record: %w", err)
	}

	return nil
}
