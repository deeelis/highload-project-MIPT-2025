package models

type AnalysisResult struct {
	IsApproved   bool    `json:"is_approved"`
	IsSpam       bool    `json:"is_spam"`
	HasSensitive bool    `json:"has_sensitive"`
	Sentiment    float64 `json:"sentiment"`
	Language     string  `json:"language"`
}
