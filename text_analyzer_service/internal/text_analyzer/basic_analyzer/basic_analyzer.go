package usecases

import (
	"strings"
	"text_analyzer_service/internal/domain/models"
)

type BasicTextAnalyzer struct{}

func NewBasicTextAnalyzer() *BasicTextAnalyzer {
	return &BasicTextAnalyzer{}
}

func (a *BasicTextAnalyzer) Analyze(text *models.TextContent) (*models.AnalysisResult, error) {
	result := &models.AnalysisResult{
		IsApproved:   true,
		IsSpam:       false,
		HasSensitive: containsSensitive(text.Data),
		Sentiment:    0.1,
		Language:     "en",
	}

	if result.HasSensitive {
		result.IsApproved = false
	}

	return result, nil
}

func containsSensitive(text string) bool {
	sensitiveWords := []string{"bad", "sensitive", "опасный"}
	for _, word := range sensitiveWords {
		if strings.Contains(strings.ToLower(text), word) {
			return true
		}
	}
	return false
}
