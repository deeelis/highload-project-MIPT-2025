package grpc

import (
	"context"
	storage "github.com/deeelis/storage-protos/gen/go/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"storage_service/internal/config"
	"storage_service/internal/domain/models"
	"storage_service/internal/domain/repositories"
	usecases "storage_service/internal/usecases"
	"strings"
)

type StorageController struct {
	cfg *config.Config
	log *slog.Logger
	storage.UnimplementedStorageServiceServer
	storageUsecase repositories.StorageUsecase
}

func NewStorageController(cfg *config.Config, log *slog.Logger) (*StorageController, error) {
	ctx := context.Background()
	storageUsecase, err := usecases.NewStorageUsecase(ctx, cfg, log)
	if err != nil {
		return nil, err
	}

	return &StorageController{
		cfg:            cfg,
		log:            log,
		storageUsecase: storageUsecase,
	}, nil
}

func (c *StorageController) GetContent(ctx context.Context, req *storage.ContentRequest) (*storage.ContentResponse, error) {
	content, err := c.storageUsecase.GetContent(ctx, req.ContentId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "content not found: %v", err)
	}

	var processingStatus storage.ProcessingStatus
	switch content.Status {
	case models.StatusPending:
		processingStatus = storage.ProcessingStatus_PENDING
	case models.StatusProcessing:
		processingStatus = storage.ProcessingStatus_PROCESSING
	case models.StatusCompleted:
		processingStatus = storage.ProcessingStatus_COMPLETED
	case models.StatusFailed:
		processingStatus = storage.ProcessingStatus_FAILED
	default:
		return nil, status.Errorf(codes.Internal, "unknown processing status: %v", content.Status)
	}

	resp := &storage.ContentResponse{
		ContentId: content.ID,
		Type:      storage.ContentType(storage.ContentType_value[strings.ToUpper(string(content.Type))]),
		Status:    processingStatus,
	}

	switch content.Type {
	case models.ContentTypeText:
		textContent, err := c.storageUsecase.GetTextContent(ctx, content.ID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get text content: %v", err)
		}
		resp.Content = &storage.ContentResponse_Text{
			Text: &storage.TextContent{
				OriginalText:     textContent.OriginalText,
				AnalysisMetadata: textContent.Metadata,
			},
		}
		c.log.Debug(textContent.OriginalText)
	case models.ContentTypeImage:
		imageContent, err := c.storageUsecase.GetImageContent(ctx, content.ID)
		if err != nil {
			c.log.Debug("hueviy image")
			return nil, status.Errorf(codes.Internal, "failed to get image content: %v", err)
		}
		resp.Content = &storage.ContentResponse_Image{
			Image: &storage.ImageContent{
				AnalysisMetadata: imageContent.Metadata,
				ImageUrl:         imageContent.S3Key,
			},
		}
	}

	return resp, nil
}

func (c *StorageController) RegisterContent(ctx context.Context, req *storage.RegisterContentRequest) (*storage.RegisterContentResponse, error) {
	contentType := models.ContentTypeText
	if req.Type == storage.ContentType_IMAGE {
		contentType = models.ContentTypeImage
	}

	if err := c.storageUsecase.CreateContentRecord(ctx, contentType, req.ContentId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register content: %v", err)
	}

	return &storage.RegisterContentResponse{
		Success: true,
	}, nil
}
