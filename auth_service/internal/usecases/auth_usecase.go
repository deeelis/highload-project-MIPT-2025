package usecases

import (
	"auth_service/internal/domain/models"
	"context"
)

type AuthUsecase interface {
	Register(ctx context.Context, user *models.User) (string, error)
	Login(ctx context.Context, email, password string) (*models.TokenDetails, error)
	ValidateToken(token string) (string, error)
	RefreshToken(refreshToken string) (*models.TokenDetails, error)
}
