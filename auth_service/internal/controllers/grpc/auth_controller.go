package grpc

import (
	"auth_service/internal/config"
	e "auth_service/internal/domain/errors"
	"auth_service/internal/domain/models"
	"auth_service/internal/usecases"
	"auth_service/internal/usecases/auth_usecase"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	auth "github.com/deeelis/auth-protos/gen/go/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthController struct {
	cfg *config.Config
	log *slog.Logger
	auth.UnimplementedAuthServiceServer
	authUsecase usecases.AuthUsecase
}

func NewAuthController(cfg *config.Config, log *slog.Logger) (*AuthController, error) {
	const op = "grpc.AuthController.New"
	log = log.With(slog.String("op", op))
	log.Info("initializing auth controller",
		slog.String("token_ttl", cfg.Token.TokenTTL.String()),
		slog.String("refresh_token_ttl", cfg.Token.RefreshTokenTTL.String()))

	startTime := time.Now()
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Error("empty secret key")
		return nil, errors.New("empty secret key")
	}

	authUsecase, err := auth_usecase.NewAuthUsecase(cfg, log, secretKey, cfg.Token.TokenTTL, cfg.Token.RefreshTokenTTL)
	if err != nil {
		log.Error("failed to create auth usecase",
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startTime)))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("auth controller initialized successfully",
		slog.Duration("duration", time.Since(startTime)))

	return &AuthController{authUsecase: authUsecase, cfg: cfg, log: log}, nil
}

func (c *AuthController) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	const op = "grpc.AuthController.Register"
	log := c.log.With(
		slog.String("op", op),
		slog.String("email", req.Email),
		slog.Int("name_length", len(req.Name)),
	)

	log.Info("registering new user")
	startTime := time.Now()

	user := &models.User{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	userID, err := c.authUsecase.Register(ctx, user)
	if err != nil {
		if errors.Is(err, e.ErrUserAlreadyExists) {
			log.Warn("user already exists",
				slog.Duration("duration", time.Since(startTime)))
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		log.Error("registration failed",
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startTime)))
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	log.Info("user registered successfully",
		slog.String("user_id", userID),
		slog.Duration("duration", time.Since(startTime)))

	return &auth.RegisterResponse{UserId: userID}, nil
}

func (c *AuthController) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	const op = "grpc.AuthController.Login"
	log := c.log.With(
		slog.String("op", op),
		slog.String("email", req.Email),
	)

	log.Info("authenticating user")
	startTime := time.Now()

	tokens, err := c.authUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		log.Error(err.Error())
		if errors.Is(err, e.ErrInvalidCredentials) {
			log.Warn("invalid credentials",
				slog.Duration("duration", time.Since(startTime)))
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		log.Error("login failed",
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startTime)))
		return nil, status.Error(codes.Internal, "failed to login")
	}

	log.Info("user authenticated successfully",
		slog.Duration("duration", time.Since(startTime)))

	return &auth.LoginResponse{
		Token:        tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (c *AuthController) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	const op = "grpc.AuthController.ValidateToken"
	log := c.log.With(
		slog.String("op", op),
		slog.Int("token_length", len(req.Token)),
	)

	log.Debug("validating token")
	startTime := time.Now()

	userID, err := c.authUsecase.ValidateToken(req.Token)
	if err != nil {
		log.Warn("invalid token",
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startTime)))
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	log.Debug("token validated successfully",
		slog.String("user_id", userID),
		slog.Duration("duration", time.Since(startTime)))

	return &auth.ValidateTokenResponse{
		UserId: userID,
		Valid:  true,
	}, nil
}

func (c *AuthController) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error) {
	const op = "grpc.AuthController.RefreshToken"
	log := c.log.With(
		slog.String("op", op),
		slog.Int("token_length", len(req.RefreshToken)),
	)

	log.Info("refreshing token")
	startTime := time.Now()

	tokens, err := c.authUsecase.RefreshToken(req.RefreshToken)
	if err != nil {
		log.Warn("invalid refresh token",
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startTime)))
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	log.Info("token refreshed successfully",
		slog.Duration("duration", time.Since(startTime)))

	return &auth.RefreshTokenResponse{
		Token:        tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
