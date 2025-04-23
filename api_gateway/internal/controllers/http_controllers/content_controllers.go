package http_controllers

import (
	"api_gateway/internal/config"
	"api_gateway/internal/domain/errors"
	"api_gateway/internal/domain/models"
	"api_gateway/internal/usecases"
	"api_gateway/internal/usecases/content_usecase"
	"api_gateway/logger"
	"context"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type ContentController struct {
	cfg       *config.Config
	contentUC usecases.ContentUsecase
	log       *slog.Logger
}

func NewContentController(cfg *config.Config, log *slog.Logger) (*ContentController, error) {
	const op = "http_controllers.NewContentController"
	log = log.With(slog.String("op", op))
	log.Info("initializing content controller")

	uc, err := content_usecase.NewContentUseCase(context.Background(), cfg, log)
	if err != nil {
		log.Error("failed to create content use case", logger.Err(err))
		return nil, err
	}

	log.Info("content controller initialized successfully")
	return &ContentController{
		contentUC: uc,
		log:       log,
	}, nil
}

func (c *ContentController) UploadText(ctx *gin.Context) {
	const op = "http_controllers.ContentController.UploadText"
	log := c.log.With(
		slog.String("op", op),
		slog.String("method", ctx.Request.Method),
		slog.String("path", ctx.FullPath()),
	)
	userID, _ := ctx.Get("userID")
	log = log.With(slog.String("user_id", userID.(string)))

	log.Info("handling text upload request")

	var req struct {
		Text string `json:"text" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid request body",
			logger.Err(err),
			slog.Int("text_length", len(req.Text)))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	log.Debug("processing text content",
		slog.Int("text_length", len(req.Text)))

	content, err := c.contentUC.ProcessContent(userID.(string), models.ContentTypeText, req.Text, "text/plain")
	if err != nil {
		log.Error("failed to process text content",
			logger.Err(err),
			slog.String("content_type", string(models.ContentTypeText)))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	log.Info("text content processed successfully",
		slog.String("content_id", content.ID),
		slog.String("content_status", content.Status))

	ctx.JSON(http.StatusAccepted, gin.H{
		"id":     content.ID,
		"status": content.Status,
	})
}

func (c *ContentController) UploadImage(ctx *gin.Context) {
	const op = "http_controllers.ContentController.UploadImage"
	log := c.log.With(
		slog.String("op", op),
		slog.String("method", ctx.Request.Method),
		slog.String("path", ctx.FullPath()),
	)
	userID, _ := ctx.Get("userID")
	log = log.With(slog.String("user_id", userID.(string)))

	log.Info("handling image upload request")

	file, err := ctx.FormFile("image")
	if err != nil {
		log.Warn("failed to get image file from request",
			logger.Err(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid image"})
		return
	}

	log.Debug("processing image file",
		slog.String("filename", file.Filename),
		slog.Int64("size", file.Size))

	fileData, err := file.Open()
	if err != nil {
		log.Error("failed to open uploaded image file",
			logger.Err(err),
			slog.String("filename", file.Filename))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer fileData.Close()

	imageBytes, err := io.ReadAll(fileData)
	if err != nil {
		log.Error("failed to read image file data",
			logger.Err(err),
			slog.String("filename", file.Filename))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if len(imageBytes) > 1<<20 {
		log.Warn("image file size exceeds limit",
			slog.Int("size_bytes", len(imageBytes)),
			slog.Int("max_size_bytes", 1<<20))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "image file too large, max 1MB"})
		return
	}

	mimeType := http.DetectContentType(imageBytes)
	if !strings.HasPrefix(mimeType, "image/") {
		log.Warn("uploaded file is not a valid image",
			slog.String("mime_type", mimeType),
			slog.String("filename", file.Filename))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "uploaded file is not an image"})
		return
	}

	log.Debug("image file validated",
		slog.String("mime_type", mimeType),
		slog.Int("size_bytes", len(imageBytes)))

	base64Str := base64.StdEncoding.EncodeToString(imageBytes)

	content, err := c.contentUC.ProcessContent(
		userID.(string),
		models.ContentTypeImage,
		base64Str,
		mimeType,
	)
	if err != nil {
		log.Error("failed to process image content",
			logger.Err(err),
			slog.String("content_type", string(models.ContentTypeImage)))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	log.Info("image content processed successfully",
		slog.String("content_id", content.ID),
		slog.String("content_status", content.Status))

	ctx.JSON(http.StatusAccepted, gin.H{
		"id":     content.ID,
		"status": content.Status,
	})
}

func (c *ContentController) GetContent(ctx *gin.Context) {
	const op = "http_controllers.ContentController.GetContent"
	log := c.log.With(
		slog.String("op", op),
		slog.String("method", ctx.Request.Method),
		slog.String("path", ctx.FullPath()),
	)
	userID, _ := ctx.Get("userID")
	contentID := ctx.Param("id")
	log = log.With(
		slog.String("user_id", userID.(string)),
		slog.String("content_id", contentID),
	)

	log.Info("handling get content request")

	if contentID == "" {
		log.Warn("empty content id in request")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "content id is required"})
		return
	}

	contentStatus, err := c.contentUC.GetContent(ctx, contentID)
	if err != nil {
		log.Error("failed to retrieve content",
			logger.Err(err))

		switch err {
		case errors.ErrInvalidCredentials:
			log.Warn("invalid credentials for content access")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid content id"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	log.Debug("content retrieved successfully",
		slog.String("content_status", contentStatus.Status))

	content := models.Content{
		ID:     contentStatus.ID,
		UserID: userID.(string),
		Type:   models.ContentType(contentStatus.Type),
		Data:   contentStatus.OriginalContent,
		Status: contentStatus.Status,
	}

	response := gin.H{
		"id":     contentStatus.ID,
		"status": contentStatus.Status,
		"type":   contentStatus.Type,
	}

	if len(contentStatus.Analysis) > 0 {
		response["analysis"] = contentStatus.Analysis
		log.Debug("content analysis included in response",
			slog.Int("analysis_length", len(contentStatus.Analysis)))
	}

	if contentStatus.Status == "COMPLETED" {
		response["data"] = content.Data
		log.Debug("content data included in response",
			slog.Int("data_length", len(content.Data)))
	}

	log.Info("content request completed successfully")
	ctx.JSON(http.StatusOK, response)
}
