package grpc

import (
	"api_gateway/internal/config"
	"api_gateway/internal/domain/errors"
	"api_gateway/internal/domain/models"
	"api_gateway/logger"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"

	authpb "github.com/deeelis/auth-protos/gen/go/auth"
)

type AuthClient struct {
	cfg    *config.AuthConfig
	client authpb.AuthServiceClient
	conn   *grpc.ClientConn
	log    *slog.Logger
}

func NewAuthClient(cfg *config.AuthConfig, log *slog.Logger) (*AuthClient, error) {
	conn, err := grpc.Dial(
		cfg.ServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	return &AuthClient{
		client: authpb.NewAuthServiceClient(conn),
		conn:   conn,
		cfg:    cfg,
		log:    log,
	}, nil
}

func (c *AuthClient) Register(ctx context.Context, email, password, name string) (*models.TokenDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	resp, err := c.client.Register(ctx, &authpb.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
	})
	if err != nil {
		c.log.Error("auth service register failed", logger.Err(err))
		return nil, errors.ErrInternalServer
	}

	return &models.TokenDetails{
		UserID: resp.UserId,
	}, nil
}

func (c *AuthClient) Login(ctx context.Context, email, password string) (*models.TokenDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	resp, err := c.client.Login(ctx, &authpb.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		c.log.Error("auth service login failed", logger.Err(err))
		return nil, errors.ErrInvalidCredentials
	}

	return &models.TokenDetails{
		AccessToken:  resp.Token,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	resp, err := c.client.ValidateToken(ctx, &authpb.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		c.log.Error("auth service validate token failed", logger.Err(err))
		return "", errors.ErrUnauthorized
	}

	if !resp.Valid {
		return "", errors.ErrUnauthorized
	}

	return resp.UserId, nil
}

func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	resp, err := c.client.RefreshToken(ctx, &authpb.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		c.log.Error("auth service refresh token failed", logger.Err(err))
		return nil, errors.ErrUnauthorized
	}

	return &models.TokenDetails{
		AccessToken:  resp.Token,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}
