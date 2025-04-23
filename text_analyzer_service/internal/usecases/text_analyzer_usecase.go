package usecases

import (
	"context"
	"text_analyzer_service/internal/domain/models"
)

type TextAnalyzerUsecase interface {
	ProcessText(ctx context.Context, text *models.TextContent) (*models.AnalysisResult, error)
}
