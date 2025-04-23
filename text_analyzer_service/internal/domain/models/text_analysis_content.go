package models

type TextAnalysisContent struct {
	ID     string         `json:"id"`
	Data   string         `json:"data"`
	UserID string         `json:"user_id"`
	Result AnalysisResult `json:"result"`
}
