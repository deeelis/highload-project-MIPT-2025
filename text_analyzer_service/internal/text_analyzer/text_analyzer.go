package models

import "text_analyzer_service/internal/domain/models"

type TextAnalyzer interface {
	Analyze(text *models.TextContent) (*models.AnalysisResult, error)
}
