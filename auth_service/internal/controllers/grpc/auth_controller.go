package grpc

import (
	"auth_service/internal/config"
	e "auth_service/internal/domain/errors"
	"auth_service/internal/domain/models"
	"auth_service/internal/usecases"
	"auth_service/internal/usecases/auth_usecase"
	"context"
	"errors"
	"log/slog"
	"os"

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
	secretKey := os.Getenv("SECRET_KEY")

	authUsecase, err := auth_usecase.NewAuthUsecase(cfg, log, secretKey, cfg.Token.TokenTTL, cfg.Token.RefreshTokenTTL)
	if err != nil {
		log.Error("auth_usecase", err.Error())
		return nil, err
	}
	return &AuthController{authUsecase: authUsecase, cfg: cfg, log: log}, nil
}

func (c *AuthController) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	user := &models.User{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	userID, err := c.authUsecase.Register(ctx, user)
	if err != nil {
		if errors.Is(err, e.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &auth.RegisterResponse{UserId: userID}, nil
}

func (c *AuthController) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	tokens, err := c.authUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, e.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &auth.LoginResponse{
		Token:        tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (c *AuthController) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	userID, err := c.authUsecase.ValidateToken(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return &auth.ValidateTokenResponse{
		UserId: userID,
		Valid:  true,
	}, nil
}

func (c *AuthController) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error) {
	tokens, err := c.authUsecase.RefreshToken(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	return &auth.RefreshTokenResponse{
		Token:        tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
