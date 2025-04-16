package auth_usecase

import (
	"api_gateway/internal/config"
	auth_client "api_gateway/internal/grpc/auth_client"
	"context"
	"log/slog"

	"api_gateway/internal/domain/errors"
	"api_gateway/internal/domain/models"
)

type AuthUsecase struct {
	cfg    *config.AuthConfig
	client *auth_client.AuthClient
	log    *slog.Logger
}

func NewAuthUsecase(cfg *config.AuthConfig, log *slog.Logger) (*AuthUsecase, error) {
	client, err := auth_client.NewAuthClient(cfg, log)
	if err != nil {
		return nil, err
	}
	return &AuthUsecase{
		client: client,
		log:    log,
	}, nil
}

func (uc *AuthUsecase) Register(ctx context.Context, email, password, name string) (*models.TokenDetails, error) {
	if err := validateCredentials(email, password, name); err != nil {
		return nil, err
	}

	return uc.client.Register(ctx, email, password, name)
}

func (uc *AuthUsecase) Login(ctx context.Context, email, password string) (*models.TokenDetails, error) {
	if email == "" || password == "" {
		return nil, errors.ErrInvalidCredentials
	}

	return uc.client.Login(ctx, email, password)
}

func (uc *AuthUsecase) ValidateToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", errors.ErrUnauthorized
	}

	return uc.client.ValidateToken(ctx, token)
}

func (uc *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenDetails, error) {
	if refreshToken == "" {
		return nil, errors.ErrUnauthorized
	}

	return uc.client.RefreshToken(ctx, refreshToken)
}

func validateCredentials(email, password, name string) error {
	// TODO: добавить валидацию email и password
	if email == "" || password == "" || name == "" {
		return errors.ErrInvalidCredentials
	}
	return nil
}
