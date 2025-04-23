package models

import "time"

type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
)

type ProcessingStatus string

const (
	StatusPending    ProcessingStatus = "pending"
	StatusProcessing ProcessingStatus = "processing"
	StatusCompleted  ProcessingStatus = "completed"
	StatusFailed     ProcessingStatus = "failed"
)

type Content struct {
	ID        string
	UserID    string `json:"user_id"`
	Type      ContentType
	Status    ProcessingStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	Metadata  map[string]string
}

type TextContent struct {
	Content
	OriginalText string
}

type ImageContent struct {
	Content
	S3Key string
}
