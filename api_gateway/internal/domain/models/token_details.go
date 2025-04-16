package models

type TokenDetails struct {
	UserID       string `json:"user_id,omitempty"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}
