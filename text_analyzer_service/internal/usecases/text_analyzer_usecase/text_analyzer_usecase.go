package text_analyzer_usecase

import (
	"context"
	"log/slog"
	"text_analyzer_service/internal/domain/models"
	models2 "text_analyzer_service/internal/text_analyzer"
	"text_analyzer_service/internal/text_analyzer/basic_analyzer"
	"text_analyzer_service/logger"
)

type TextAnalyzerUsecase struct {
	analyzer models2.TextAnalyzer
	log      *slog.Logger
}

func NewTextAnalyzerUseCase(log *slog.Logger) *TextAnalyzerUsecase {
	a := usecases.NewBasicTextAnalyzer()
	return &TextAnalyzerUsecase{
		analyzer: a,
		log:      log,
	}
}

func (uc *TextAnalyzerUsecase) ProcessText(ctx context.Context, text *models.TextContent) (*models.AnalysisResult, error) {
	result, err := uc.analyzer.Analyze(text)
	if err != nil {
		uc.log.Error("text analysis failed",
			"text_id", text.ID,
			logger.Err(err),
		)
		return nil, err
	}

	uc.log.Info("text processed successfully",
		"text_id", text.ID,
		"is_approved", result.IsApproved,
		"has_sensitive", result.HasSensitive,
	)
	return result, nil
}
