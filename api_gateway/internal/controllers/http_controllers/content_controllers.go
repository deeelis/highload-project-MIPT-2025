package http_controllers

import (
	"api_gateway/internal/config"
	"api_gateway/internal/domain/errors"
	"api_gateway/internal/domain/models"
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
	contentUC *content_usecase.UseCase
	log       *slog.Logger
}

func NewContentController(cfg *config.Config, log *slog.Logger) (*ContentController, error) {
	uc, err := content_usecase.NewContentUseCase(context.Background(), cfg, log)
	if err != nil {
		return nil, err
	}
	return &ContentController{
		contentUC: uc,
		log:       log,
	}, nil
}

func (c *ContentController) UploadText(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	var req struct {
		Text string `json:"text" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Warn("invalid request body", logger.Err(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	content, err := c.contentUC.ProcessContent(userID.(string), models.ContentTypeText, req.Text, "text/plain")
	if err != nil {
		c.log.Error("failed to process text", logger.Err(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"id":     content.ID,
		"status": content.Status,
	})
}

func (c *ContentController) UploadImage(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	file, err := ctx.FormFile("image")
	if err != nil {
		c.log.Warn("failed to get image file", logger.Err(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid image"})
		return
	}

	fileData, err := file.Open()
	if err != nil {
		c.log.Error("failed to open image file", logger.Err(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer fileData.Close()

	imageBytes, err := io.ReadAll(fileData)
	if err != nil {
		c.log.Error("failed to read image file", logger.Err(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if len(imageBytes) > 10<<20 {
		c.log.Warn("image file too large")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "image file too large, max 10MB"})
		return
	}

	mimeType := http.DetectContentType(imageBytes)
	if !strings.HasPrefix(mimeType, "image/") {
		c.log.Warn("uploaded file is not an image", slog.String("mime_type", mimeType))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "uploaded file is not an image"})
		return
	}

	base64Str := base64.StdEncoding.EncodeToString(imageBytes)

	content, err := c.contentUC.ProcessContent(
		userID.(string),
		models.ContentTypeImage,
		base64Str,
		mimeType,
	)
	if err != nil {
		c.log.Error("failed to process image", logger.Err(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"id":     content.ID,
		"status": content.Status,
	})
}

func (c *ContentController) GetContent(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	contentID := ctx.Param("id")
	if contentID == "" {
		c.log.Warn("empty content id")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "content id is required"})
		return
	}

	contentStatus, err := c.contentUC.GetContent(ctx, contentID)
	if err != nil {
		c.log.Error("failed to get content",
			logger.Err(err),
			slog.String("content_id", contentID))

		switch err {
		case errors.ErrInvalidCredentials:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid content id"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

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
	}

	if contentStatus.Status == "COMPLETED" {
		response["data"] = content.Data
	}

	ctx.JSON(http.StatusOK, response)
}
