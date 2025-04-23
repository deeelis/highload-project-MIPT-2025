package models

type TextMessage struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	UserID   string            `json:"user_id"`
	Analysis map[string]string `json:"result"`
}

type ImageMessage struct {
	ID         string           `json:"id"`
	UserID     string           `json:"user_id"`
	Data       string           `json:"data"`
	NsfwScores NsfwScoresResult `json:"nsfw_scores"`
}

type ImageKafkaMessage struct {
	ID         string           `json:"id"`
	UserID     string           `json:"user_id"`
	Data       string           `json:"data"`
	NsfwScores NsfwScoresResult `json:"nsfw_scores"`
	IsNsfw     bool             `json:"is_nsfw"`
}

type NsfwScoresResult struct {
	Drawings float64 `json:"drawings"`
	Hentai   float64 `json:"hentai"`
	Neutral  float64 `json:"neutral"`
	Porn     float64 `json:"porn"`
	Sexy     float64 `json:"sexy"`
}

type TextKafkaMessage struct {
	ID       string         `json:"id"`
	Data     string         `json:"data"`
	UserID   string         `json:"user_id"`
	Analysis AnalysisResult `json:"result"`
}

type AnalysisResult struct {
	IsApproved   bool    `json:"is_approved"`
	IsSpam       bool    `json:"is_spam"`
	HasSensitive bool    `json:"has_sensitive"`
	Sentiment    float64 `json:"sentiment"`
	Language     string  `json:"language"`
}
