package auth_usecase

import (
	"auth_service/internal/config"
	e "auth_service/internal/domain/errors"
	"auth_service/internal/domain/models"
	"auth_service/internal/repositories"
	"auth_service/internal/repositories/postgres"
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type AuthUsecase struct {
	cfg           *config.Config
	log           *slog.Logger
	userRepo      repositories.UserRepository
	secretKey     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewAuthUsecase(
	cfg *config.Config,
	log *slog.Logger,
	secretKey string,
	accessExpiry time.Duration,
	refreshExpiry time.Duration,
) (*AuthUsecase, error) {
	const op = "auth_usecase.New"
	log = log.With(slog.String("op", op))

	log.Info("initializing auth usecase",
		slog.String("access_expiry", accessExpiry.String()),
		slog.String("refresh_expiry", refreshExpiry.String()))

	startTime := time.Now()
	defer func() {
		log.Info("auth usecase initialization completed",
			slog.Duration("duration", time.Since(startTime)))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userRepo, err := postgres.NewUserRepository(ctx, &cfg.Database, log)
	if err != nil {
		log.Error("failed to create user repository",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("auth usecase initialized successfully")
	return &AuthUsecase{
		cfg:           cfg,
		log:           log,
		userRepo:      userRepo,
		secretKey:     secretKey,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}, nil
}

func (uc *AuthUsecase) Register(ctx context.Context, user *models.User) (string, error) {
	const op = "auth_usecase.Register"
	log := uc.log.With(
		slog.String("op", op),
	)

	log.Info("registering new user")
	startTime := time.Now()

	existingUser, err := uc.userRepo.GetByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, e.ErrUserNotFound) {
		log.Error("failed to check user existence",
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startTime)))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if existingUser != nil {
		log.Warn("user already exists",
			slog.Duration("duration", time.Since(startTime)))
		return "", e.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password",
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startTime)))
		return "", fmt.Errorf("%s: %w", op, err)
	}
	user.Password = string(hashedPassword)

	userID, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		log.Error("failed to create user in repository",
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startTime)))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered successfully",
		slog.String("user_id", userID),
		slog.Duration("duration", time.Since(startTime)))
	return userID, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, email, password string) (*models.TokenDetails, error) {
	const op = "auth_usecase.Login"
	log := uc.log.With(
		slog.String("op", op),
	)

	log.Info("authenticating user")
	startTime := time.Now()

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, e.ErrUserNotFound) {
			log.Warn("user not found",
				slog.Duration("duration", time.Since(startTime)))
		} else {
			log.Error("failed to retrieve user",
				slog.String("error", err.Error()),
				slog.Duration("duration", time.Since(startTime)))
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Warn("invalid password provided",
			slog.Duration("duration", time.Since(startTime)))
		return nil, e.ErrInvalidCredentials
	}

	tokens, err := uc.generateTokens(user.ID)
	if err != nil {
		log.Error("failed to generate tokens",
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startTime)))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user authenticated successfully",
		slog.String("user_id", user.ID),
		slog.Duration("duration", time.Since(startTime)))

	return tokens, nil
}

func (uc *AuthUsecase) ValidateToken(tokenString string) (string, error) {
	const op = "auth_usecase.ValidateToken"
	log := uc.log.With(
		slog.String("op", op),
		slog.Int("token_length", len(tokenString)),
	)

	log.Debug("validating token")
	startTime := time.Now()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Warn("unexpected token signing method",
				slog.String("alg", token.Header["alg"].(string)))
			return nil, e.ErrInvalidToken
		}
		return []byte(uc.secretKey), nil
	})

	if err != nil {
		var jwtErr *jwt.ValidationError
		if errors.As(err, &jwtErr) {
			log.Warn("token validation failed",
				slog.String("error", jwtErr.Error()),
				slog.Duration("duration", time.Since(startTime)))
		} else {
			log.Error("failed to parse token",
				slog.String("error", err.Error()),
				slog.Duration("duration", time.Since(startTime)))
		}
		return "", e.ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			log.Warn("missing user_id in token claims",
				slog.Duration("duration", time.Since(startTime)))
			return "", e.ErrInvalidToken
		}

		log.Debug("token validated successfully",
			slog.String("user_id", userID),
			slog.Duration("duration", time.Since(startTime)))

		return userID, nil
	}

	log.Warn("invalid token claims",
		slog.Duration("duration", time.Since(startTime)))
	return "", e.ErrInvalidToken
}

func (uc *AuthUsecase) RefreshToken(refreshToken string) (*models.TokenDetails, error) {
	const op = "auth_usecase.RefreshToken"
	log := uc.log.With(
		slog.String("op", op),
		slog.Int("token_length", len(refreshToken)),
	)

	log.Info("refreshing token")
	startTime := time.Now()

	userID, err := uc.ValidateToken(refreshToken)
	if err != nil {
		log.Warn("invalid refresh token provided",
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startTime)))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tokens, err := uc.generateTokens(userID)
	if err != nil {
		log.Error("failed to generate new tokens",
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startTime)))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("tokens refreshed successfully",
		slog.String("user_id", userID),
		slog.Duration("duration", time.Since(startTime)))

	return tokens, nil
}

func (uc *AuthUsecase) generateTokens(userID string) (*models.TokenDetails, error) {
	const op = "auth_usecase.generateTokens"
	log := uc.log.With(
		slog.String("op", op),
		slog.String("user_id", userID),
	)

	log.Debug("generating new tokens")
	startTime := time.Now()

	now := time.Now()
	accessExpire := now.Add(uc.accessExpiry).Unix()
	refreshExpire := now.Add(uc.refreshExpiry).Unix()

	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     accessExpire,
		"iat":     now.Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(uc.secretKey))
	if err != nil {
		log.Error("failed to sign access token",
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startTime)))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     refreshExpire,
		"iat":     now.Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(uc.secretKey))
	if err != nil {
		log.Error("failed to sign refresh token",
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startTime)))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("tokens generated successfully",
		slog.Int64("access_expire", accessExpire),
		slog.Int64("refresh_expire", refreshExpire),
		slog.Duration("duration", time.Since(startTime)))

	return &models.TokenDetails{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		AtExpires:    accessExpire,
		RtExpires:    refreshExpire,
	}, nil
}
