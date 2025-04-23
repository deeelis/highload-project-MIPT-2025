package grpc

import (
	"api_gateway/internal/config"
	"api_gateway/internal/domain/errors"
	"api_gateway/internal/domain/models"
	"api_gateway/logger"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"log/slog"
	"time"

	authpb "github.com/deeelis/auth-protos/gen/go/auth"
)

type AuthClient struct {
	cfg    *config.AuthConfig
	client authpb.AuthServiceClient
	conn   *grpc.ClientConn
	log    *slog.Logger
}

func NewAuthClient(cfg *config.AuthConfig, log *slog.Logger) (*AuthClient, error) {
	const op = "grpc.AuthClient.New"
	log = log.With(slog.String("op", op))
	log.Info("initializing auth gRPC client",
		slog.String("address", cfg.ServiceAddress),
		slog.Duration("timeout", cfg.Timeout))

	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		cfg.ServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Error("failed to connect to auth service",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	log.Info("auth gRPC client initialized successfully",
		slog.Duration("duration", time.Since(startTime)))

	return &AuthClient{
		client: authpb.NewAuthServiceClient(conn),
		conn:   conn,
		cfg:    cfg,
		log:    log,
	}, nil
}

func (c *AuthClient) Register(ctx context.Context, email, password, name string) (*models.TokenDetails, error) {
	const op = "grpc.AuthClient.Register"
	log := c.log.With(
		slog.String("op", op),
		slog.String("email", email),
		slog.Int("name_length", len(name)),
	)

	log.Info("registering new user")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	resp, err := c.client.Register(ctx, &authpb.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
	})
	if err != nil {
		grpcStatus, _ := status.FromError(err)
		log.Error("registration failed",
			logger.Err(err),
			slog.String("grpc_code", grpcStatus.Code().String()),
			slog.Duration("duration", time.Since(startTime)))
		return nil, mapGRPCError(err)
	}

	log.Info("user registered successfully",
		slog.String("user_id", resp.UserId),
		slog.Duration("duration", time.Since(startTime)))

	return &models.TokenDetails{
		UserID: resp.UserId,
	}, nil
}

func (c *AuthClient) Login(ctx context.Context, email, password string) (*models.TokenDetails, error) {
	const op = "grpc.AuthClient.Login"
	log := c.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("authenticating user")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	resp, err := c.client.Login(ctx, &authpb.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		grpcStatus, _ := status.FromError(err)
		log.Error("login failed",
			logger.Err(err),
			slog.String("grpc_code", grpcStatus.Code().String()),
			slog.Duration("duration", time.Since(startTime)))
		return nil, mapGRPCError(err)
	}

	log.Info("user authenticated successfully",
		slog.Duration("duration", time.Since(startTime)))

	return &models.TokenDetails{
		AccessToken:  resp.Token,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (string, error) {
	const op = "grpc.AuthClient.ValidateToken"
	log := c.log.With(
		slog.String("op", op),
		slog.Int("token_length", len(token)),
	)

	log.Debug("validating token")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	resp, err := c.client.ValidateToken(ctx, &authpb.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		grpcStatus, _ := status.FromError(err)
		log.Error("token validation failed",
			logger.Err(err),
			slog.String("grpc_code", grpcStatus.Code().String()),
			slog.Duration("duration", time.Since(startTime)))
		return "", mapGRPCError(err)
	}

	if !resp.Valid {
		log.Warn("invalid token provided",
			slog.Duration("duration", time.Since(startTime)))
		return "", errors.ErrUnauthorized
	}

	log.Debug("token validated successfully",
		slog.String("user_id", resp.UserId),
		slog.Duration("duration", time.Since(startTime)))

	return resp.UserId, nil
}

func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenDetails, error) {
	const op = "grpc.AuthClient.RefreshToken"
	log := c.log.With(
		slog.String("op", op),
		slog.Int("token_length", len(refreshToken)),
	)

	log.Info("refreshing token")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	resp, err := c.client.RefreshToken(ctx, &authpb.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		grpcStatus, _ := status.FromError(err)
		log.Error("token refresh failed",
			logger.Err(err),
			slog.String("grpc_code", grpcStatus.Code().String()),
			slog.Duration("duration", time.Since(startTime)))
		return nil, mapGRPCError(err)
	}

	log.Info("token refreshed successfully",
		slog.Duration("duration", time.Since(startTime)))

	return &models.TokenDetails{
		AccessToken:  resp.Token,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (c *AuthClient) Close() error {
	const op = "grpc.AuthClient.Close"
	log := c.log.With(slog.String("op", op))

	log.Info("closing gRPC connection")
	if err := c.conn.Close(); err != nil {
		log.Error("failed to close gRPC connection", logger.Err(err))
		return err
	}

	log.Info("gRPC connection closed successfully")
	return nil
}

func mapGRPCError(err error) error {
	grpcStatus, _ := status.FromError(err)
	switch grpcStatus.Code() {
	case codes.NotFound:
		return errors.ErrUserNotFound
	case codes.AlreadyExists:
		return errors.ErrUserAlreadyExists
	case codes.Unauthenticated:
		return errors.ErrInvalidCredentials
	case codes.InvalidArgument:
		return errors.ErrInvalidInput
	default:
		return errors.ErrInternalServer
	}
}
