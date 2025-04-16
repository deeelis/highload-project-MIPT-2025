package usecases

import (
	"api_gateway/internal/domain/models"
	"context"
)

type AuthUsecase interface {
	Register(ctx context.Context, email, password, name string) (*models.TokenDetails, error)
	Login(ctx context.Context, email, password string) (*models.TokenDetails, error)
	ValidateToken(ctx context.Context, token string) (string, error)
	RefreshToken(ctx context.Context, refreshToken string) (*models.TokenDetails, error)
}
