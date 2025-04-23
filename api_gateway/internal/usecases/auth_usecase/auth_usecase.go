package auth_usecase

import (
	"api_gateway/internal/config"
	"api_gateway/internal/grpc"
	auth_client "api_gateway/internal/grpc/auth_client"
	"api_gateway/logger"
	"context"
	"log/slog"
	"time"

	"api_gateway/internal/domain/errors"
	"api_gateway/internal/domain/models"
)

type AuthUsecase struct {
	cfg    *config.AuthConfig
	client grpc.AuthClient
	log    *slog.Logger
}

func NewAuthUsecase(cfg *config.AuthConfig, log *slog.Logger) (*AuthUsecase, error) {
	const op = "auth_usecase.NewAuthUsecase"
	log = log.With(slog.String("op", op))
	log.Info("initializing auth usecase")

	startTime := time.Now()
	client, err := auth_client.NewAuthClient(cfg, log)
	if err != nil {
		log.Error("failed to create auth client",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return nil, err
	}

	log.Info("auth usecase initialized successfully",
		slog.Duration("duration", time.Since(startTime)))

	return &AuthUsecase{
		cfg:    cfg,
		client: client,
		log:    log,
	}, nil
}

func (uc *AuthUsecase) Register(ctx context.Context, email, password, name string) (*models.TokenDetails, error) {
	const op = "auth_usecase.Register"
	log := uc.log.With(
		slog.String("op", op),
		slog.String("email", email),
		slog.Int("name_length", len(name)),
	)

	log.Info("registering new user")
	startTime := time.Now()
	if err := validateCredentials(email, password, name); err != nil {
		log.Warn("invalid registration credentials",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return nil, err
	}

	tokenDetails, err := uc.client.Register(ctx, email, password, name)
	if err != nil {
		log.Error("registration failed",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return nil, err
	}

	log.Info("user registered successfully",
		slog.String("user_id", tokenDetails.UserID),
		slog.Duration("duration", time.Since(startTime)))

	return tokenDetails, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, email, password string) (*models.TokenDetails, error) {
	const op = "auth_usecase.Login"
	log := uc.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("authenticating user")
	startTime := time.Now()

	if email == "" || password == "" {
		log.Warn("empty credentials provided",
			slog.Duration("duration", time.Since(startTime)))
		return nil, errors.ErrInvalidCredentials
	}

	tokenDetails, err := uc.client.Login(ctx, email, password)
	if err != nil {
		log.Error("login failed",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return nil, err
	}

	log.Info("user authenticated successfully",
		slog.Duration("duration", time.Since(startTime)))

	return tokenDetails, nil
}

func (uc *AuthUsecase) ValidateToken(ctx context.Context, token string) (string, error) {
	const op = "auth_usecase.ValidateToken"
	log := uc.log.With(
		slog.String("op", op),
		slog.Int("token_length", len(token)),
	)

	log.Debug("validating token")
	startTime := time.Now()

	if token == "" {
		log.Warn("empty token provided",
			slog.Duration("duration", time.Since(startTime)))
		return "", errors.ErrUnauthorized
	}

	userID, err := uc.client.ValidateToken(ctx, token)
	if err != nil {
		log.Error("token validation failed",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return "", err
	}

	log.Debug("token validated successfully",
		slog.String("user_id", userID),
		slog.Duration("duration", time.Since(startTime)))

	return userID, nil
}

func (uc *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenDetails, error) {
	const op = "auth_usecase.RefreshToken"
	log := uc.log.With(
		slog.String("op", op),
		slog.Int("token_length", len(refreshToken)),
	)

	log.Info("refreshing token")
	startTime := time.Now()

	if refreshToken == "" {
		log.Warn("empty refresh token provided",
			slog.Duration("duration", time.Since(startTime)))
		return nil, errors.ErrUnauthorized
	}

	tokenDetails, err := uc.client.RefreshToken(ctx, refreshToken)
	if err != nil {
		log.Error("token refresh failed",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return nil, err
	}

	log.Info("token refreshed successfully",
		slog.Duration("duration", time.Since(startTime)))

	return tokenDetails, nil
}

func validateCredentials(email, password, name string) error {
	// TODO: добавить валидацию email и password
	if email == "" || password == "" || name == "" {
		return errors.ErrInvalidCredentials
	}
	return nil
}
