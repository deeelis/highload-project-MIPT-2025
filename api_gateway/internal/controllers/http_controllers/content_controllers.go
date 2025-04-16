package http_controllers

import (
	"api_gateway/internal/config"
	e "api_gateway/internal/domain/errors"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"api_gateway/internal/domain/models"
	"api_gateway/internal/usecases/content_usecase"
	"api_gateway/logger"
	"github.com/gin-gonic/gin"
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

	content, err := c.contentUC.ProcessContent(userID.(string), models.ContentTypeText, req.Text)
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

	// Чтение файла и конвертация в base64
	fileData, err := file.Open()
	if err != nil {
		c.log.Error("failed to open image file", logger.Err(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer fileData.Close()

	// Здесь должна быть реализация конвертации в base64
	// Для примера используем пустую строку
	content, err := c.contentUC.ProcessContent(userID.(string), models.ContentTypeImage, "")
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

func (c *ContentController) GetStatus(ctx *gin.Context) {
	contentID := ctx.Param("id")
	if contentID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "content ID is required"})
		return
	}

	status, err := c.contentUC.GetContentStatus(ctx.Request.Context(), contentID)
	if err != nil {
		c.log.Error("failed to get content status", logger.Err(err))
		if errors.Is(err, e.ErrContentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "content not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	response := gin.H{
		"id":     status.ID,
		"status": status.Status,
	}

	if status.Analysis != nil {
		response["analysis"] = status.Analysis
	}

	ctx.JSON(http.StatusOK, response)
}
