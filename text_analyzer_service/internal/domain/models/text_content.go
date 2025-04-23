package models

type TextContent struct {
	ID     string `json:"id"`
	Data   string `json:"data"`
	UserID string `json:"user_id"`
}
