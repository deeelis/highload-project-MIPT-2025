package models

type ContentStatus struct {
	ID              string
	Type            string
	Status          string
	OriginalContent string
	Analysis        map[string]interface{}
}
