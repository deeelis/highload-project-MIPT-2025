package models

import "time"

type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
)

type Content struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Type      ContentType `json:"type"`
	Data      string      `json:"data"`
	DataType  string      `json:"data_type"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
